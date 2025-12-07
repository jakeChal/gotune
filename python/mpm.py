"""
McLeod Pitch Method (MPM) implementation.

Based on the paper:
"A Smarter Way to Find Pitch" by Philip McLeod and Geoff Wyvill (2002)

The MPM algorithm uses a modified autocorrelation function called NSDF
(Normalized Square Difference Function) to detect pitch with high accuracy.
"""

import numpy as np


def normalized_square_difference(audio_buffer):
    """
    Calculate the Normalized Square Difference Function (NSDF).

    The NSDF is a key component of MPM. It's similar to autocorrelation
    but normalized to account for varying signal energy at different lags.

    NSDF formula:
    nsdf(tau) = 2 * r(tau) / [m(0) + m(tau)]

    where:
    - r(tau) is the autocorrelation at lag tau
    - m(tau) is the sum of squared samples in the shifted windows

    Parameters
    ----------
    audio_buffer : np.ndarray
        Audio signal buffer (1D array)

    Returns
    -------
    nsdf : np.ndarray
        Normalized square difference function values
    """
    n = len(audio_buffer)

    # Calculate autocorrelation using FFT (efficient O(N log N))
    # Pad to next power of 2 for FFT efficiency
    fft_size = 2 ** int(np.ceil(np.log2(2 * n)))

    # Compute autocorrelation via FFT
    fft = np.fft.rfft(audio_buffer, n=fft_size)
    autocorr = np.fft.irfft(fft * np.conj(fft))[:n]

    # Calculate m(tau) = sum of squared samples for each lag
    # m(tau) = sum(x[j]^2) + sum(x[j+tau]^2) for j in valid range
    cumsum_sq = np.concatenate(([0], np.cumsum(audio_buffer ** 2)))

    m = np.zeros(n)
    for tau in range(n):
        # m(tau) = sum(x[0:n-tau]^2) + sum(x[tau:n]^2)
        m[tau] = cumsum_sq[n - tau] + (cumsum_sq[n] - cumsum_sq[tau])

    # Avoid division by zero
    m[m == 0] = 1e-10

    # Calculate NSDF: nsdf(tau) = 2 * r(tau) / m(tau)
    nsdf = 2 * autocorr / m

    return nsdf


def peak_picking(nsdf, threshold=0.1):
    """
    Find positive peaks in the NSDF that exceed the threshold.

    A peak is defined as a point where:
    1. The value is positive
    2. The value exceeds the threshold
    3. The value is greater than its neighbors (local maximum)

    Parameters
    ----------
    nsdf : np.ndarray
        Normalized square difference function
    threshold : float
        Minimum threshold for peak detection (0 to 1)

    Returns
    -------
    peaks : list of tuples
        List of (index, value) for each detected peak
    """
    peaks = []

    # Start from lag=1 to avoid the trivial peak at lag=0
    for i in range(1, len(nsdf) - 1):
        # Check if it's a positive peak above threshold
        if nsdf[i] > threshold and nsdf[i] > 0:
            # Check if it's a local maximum
            if nsdf[i] > nsdf[i - 1] and nsdf[i] > nsdf[i + 1]:
                peaks.append((i, nsdf[i]))

    return peaks


def parabolic_interpolation(nsdf, peak_index):
    """
    Use parabolic interpolation to refine the peak location.

    This improves frequency resolution by fitting a parabola through
    the peak and its neighbors to find the true maximum.

    Parameters
    ----------
    nsdf : np.ndarray
        Normalized square difference function
    peak_index : int
        Index of the peak

    Returns
    -------
    refined_index : float
        Refined peak location (can be fractional)
    """
    if peak_index <= 0 or peak_index >= len(nsdf) - 1:
        return float(peak_index)

    # Get the three points: before, at, and after the peak
    alpha = nsdf[peak_index - 1]
    beta = nsdf[peak_index]
    gamma = nsdf[peak_index + 1]

    # Parabolic interpolation formula
    # The peak of the parabola is at: x = 0.5 * (alpha - gamma) / (alpha - 2*beta + gamma)
    denominator = alpha - 2 * beta + gamma

    if abs(denominator) < 1e-10:
        return float(peak_index)

    offset = 0.5 * (alpha - gamma) / denominator

    return peak_index + offset


def mpm_pitch_detection(audio_buffer, sample_rate, threshold=0.1):
    """
    Detect pitch using the McLeod Pitch Method (MPM).

    Algorithm steps:
    1. Calculate NSDF (Normalized Square Difference Function)
    2. Find positive peaks above threshold
    3. Select the highest peak (best candidate)
    4. Use parabolic interpolation for sub-sample accuracy
    5. Convert lag to frequency

    Parameters
    ----------
    audio_buffer : np.ndarray
        Audio signal buffer (1D array)
    sample_rate : int
        Sample rate in Hz
    threshold : float, optional
        Detection threshold (0 to 1), default 0.1
        Higher = more conservative (fewer false positives)
        Lower = more sensitive (may detect noise)

    Returns
    -------
    frequency : float or None
        Detected frequency in Hz, or None if no pitch found
    clarity : float or None
        Confidence/clarity of detection (0 to 1), or None if no pitch
    """
    # Step 1: Calculate NSDF
    nsdf = normalized_square_difference(audio_buffer)

    # Step 2: Find peaks above threshold
    peaks = peak_picking(nsdf, threshold)

    if not peaks:
        return None, None

    # Step 3: Select the best peak
    # For musical signals, the first peak (lowest lag) with high clarity
    # is usually the fundamental frequency. Harmonics appear at higher lags.
    first_peak_index, first_peak_clarity = peaks[0]

    # If first peak has strong clarity, use it (fundamental frequency)
    # Otherwise, find the peak with maximum clarity
    strong_clarity_threshold = 0.8
    if first_peak_clarity >= strong_clarity_threshold:
        best_peak = peaks[0]
    else:
        best_peak = max(peaks, key=lambda p: p[1])

    peak_index, clarity = best_peak

    # Step 4: Refine peak location with parabolic interpolation
    refined_lag = parabolic_interpolation(nsdf, peak_index)

    # Step 5: Convert lag to frequency
    # frequency = sample_rate / period (in samples)
    if refined_lag > 0:
        frequency = sample_rate / refined_lag
    else:
        return None, None

    return frequency, clarity


def detect_pitch_mpm(audio, sample_rate, frame_length=2048, hop_length=512,
                     threshold=0.1, fmin=50, fmax=2000):
    """
    Detect pitch over time using MPM with a sliding window.

    This function processes the audio in overlapping frames and
    detects pitch for each frame independently.

    Parameters
    ----------
    audio : np.ndarray
        Audio signal (1D array)
    sample_rate : int
        Sample rate in Hz
    frame_length : int
        Length of each analysis frame in samples
    hop_length : int
        Number of samples between successive frames
    threshold : float
        MPM detection threshold (0 to 1)
    fmin : float
        Minimum frequency to consider (Hz)
    fmax : float
        Maximum frequency to consider (Hz)

    Returns
    -------
    pitches : np.ndarray
        Detected pitch values in Hz for each frame (NaN if no pitch)
    clarities : np.ndarray
        Clarity/confidence values for each frame (NaN if no pitch)
    times : np.ndarray
        Time stamps for each frame in seconds
    """
    # Calculate number of frames
    num_frames = 1 + (len(audio) - frame_length) // hop_length

    pitches = np.full(num_frames, np.nan)
    clarities = np.full(num_frames, np.nan)
    times = np.zeros(num_frames)

    for i in range(num_frames):
        # Extract frame
        start = i * hop_length
        end = start + frame_length
        frame = audio[start:end]

        # Calculate time stamp for this frame
        times[i] = start / sample_rate

        # Detect pitch for this frame
        freq, clarity = mpm_pitch_detection(frame, sample_rate, threshold)

        # Filter based on frequency range
        if freq is not None:
            if fmin <= freq <= fmax:
                pitches[i] = freq
                clarities[i] = clarity

    return pitches, clarities, times


def get_dominant_pitch_mpm(pitches, clarities=None, min_clarity=0.5):
    """
    Get the dominant (most reliable) pitch from a series of detections.

    Uses weighted median based on clarity scores for robustness.

    Parameters
    ----------
    pitches : np.ndarray
        Array of pitch estimates in Hz
    clarities : np.ndarray, optional
        Clarity/confidence scores for each pitch
    min_clarity : float
        Minimum clarity threshold to consider

    Returns
    -------
    dominant_pitch : float or None
        The dominant pitch in Hz, or None if no valid pitch found
    """
    # Filter out NaN values
    valid_mask = ~np.isnan(pitches)

    if clarities is not None:
        # Also filter by clarity threshold
        valid_mask &= (clarities >= min_clarity) & ~np.isnan(clarities)

    valid_pitches = pitches[valid_mask]

    if len(valid_pitches) == 0:
        return None

    # Use median for robustness against outliers
    # (Could also use weighted median based on clarity if available)
    if clarities is not None:
        valid_clarities = clarities[valid_mask]
        # Weighted median: sort by pitch, use cumulative clarity weights
        sort_idx = np.argsort(valid_pitches)
        sorted_pitches = valid_pitches[sort_idx]
        sorted_weights = valid_clarities[sort_idx]
        cumsum = np.cumsum(sorted_weights)
        median_idx = np.searchsorted(cumsum, cumsum[-1] / 2)
        return sorted_pitches[median_idx]
    else:
        return np.median(valid_pitches)
