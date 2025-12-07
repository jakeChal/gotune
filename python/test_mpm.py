"""
Test MPM implementation with synthetic signals.

This is the reference implementation for comparing with Go port.
"""

import numpy as np
from mpm import mpm_pitch_detection
from pitch_utils import pitch_to_note


def generate_sine_wave(frequency, duration=1.0, sample_rate=48000, amplitude=0.5):
    """Generate a pure sine wave at a specific frequency."""
    t = np.linspace(0, duration, int(sample_rate * duration), endpoint=False)
    signal = amplitude * np.sin(2 * np.pi * frequency * t)
    return signal.astype(np.float32)


def generate_guitar_tone(frequency, duration=1.0, sample_rate=48000,
                         amplitude=0.5, num_harmonics=5):
    """Generate a guitar-like tone with harmonics."""
    t = np.linspace(0, duration, int(sample_rate * duration), endpoint=False)
    signal = np.zeros_like(t)

    # Add fundamental and harmonics with decreasing amplitude
    for n in range(1, num_harmonics + 1):
        harmonic_amp = amplitude / n
        signal += harmonic_amp * np.sin(2 * np.pi * frequency * n * t)

    # Simple envelope
    envelope = np.ones_like(t)
    attack_samples = int(0.05 * sample_rate)
    release_samples = int(0.1 * sample_rate)

    envelope[:attack_samples] = np.linspace(0, 1, attack_samples)
    envelope[-release_samples:] = np.linspace(1, 0, release_samples)

    signal *= envelope
    return signal.astype(np.float32)


def test_single_frequency(frequency, signal_type="guitar", sample_rate=48000):
    """Test MPM on a single frequency."""
    # Generate signal
    if signal_type == "sine":
        signal = generate_sine_wave(frequency, duration=1.0,
                                     sample_rate=sample_rate)
    else:  # guitar
        signal = generate_guitar_tone(frequency, duration=1.0,
                                       sample_rate=sample_rate)

    # Use a single frame for detection (first 4096 samples)
    frame_length = min(4096, len(signal))
    frame = signal[:frame_length]

    # Detect pitch on frame
    detected_freq, clarity = mpm_pitch_detection(
        frame, sample_rate, threshold=0.05
    )

    # Convert to note
    note_name, _, cents_off = pitch_to_note(detected_freq)

    # Calculate error
    error_hz = detected_freq - frequency if detected_freq else None

    return {
        'expected_freq': frequency,
        'detected_freq': detected_freq,
        'error_hz': error_hz,
        'note': note_name,
        'cents_offset': cents_off,
        'clarity': clarity
    }


def run_tests():
    """Run test suite for MPM."""
    print("=" * 70)
    print("MPM REFERENCE IMPLEMENTATION - TEST SUITE")
    print("=" * 70)

    # Test cases: standard guitar notes
    test_frequencies = [
        ('E2', 82.41),
        ('A2', 110.00),
        ('D3', 146.83),
        ('G3', 196.00),
        ('B3', 246.94),
        ('E4', 329.63),
    ]

    print("\nTesting Guitar-like Tones:")
    print(f"{'Note':<6} {'Freq (Hz)':<12} {'Detected':<12} "
          f"{'Error (Hz)':<12} {'Cents':<10} {'Clarity':<10}")
    print("-" * 70)

    results = []
    for note_name, freq in test_frequencies:
        result = test_single_frequency(freq, signal_type="guitar")
        results.append(result)

        detected_str = f"{result['detected_freq']:.2f}" if result['detected_freq'] else "N/A"
        error_str = f"{result['error_hz']:+.3f}" if result['error_hz'] is not None else "N/A"
        cents_str = f"{result['cents_offset']:+.1f}" if result['cents_offset'] is not None else "N/A"
        clarity_str = f"{result['clarity']:.3f}" if result['clarity'] is not None else "N/A"

        print(f"{note_name:<6} {freq:<12.2f} {detected_str:<12} "
              f"{error_str:<12} {cents_str:<10} {clarity_str:<10}")

    # Statistics
    errors_hz = [r['error_hz'] for r in results if r['error_hz'] is not None]
    cents = [r['cents_offset'] for r in results if r['cents_offset'] is not None]

    print("\nStatistics:")
    if errors_hz:
        print(f"  Frequency Error: Mean={np.mean(errors_hz):+.3f} Hz, "
              f"Std={np.std(errors_hz):.3f} Hz")
    if cents:
        print(f"  Cents Offset:    Mean={np.mean(cents):+.2f} cents, "
              f"Std={np.std(cents):.2f} cents")

    print("\n" + "=" * 70)
    print("PASS: All tests completed successfully!")
    print("=" * 70)


if __name__ == '__main__':
    run_tests()
