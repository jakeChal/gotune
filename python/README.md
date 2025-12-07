# MPM Reference Implementation

This directory contains the **Python reference implementation** of the McLeod Pitch Method (MPM) for porting to Go.

## Contents

- **`mpm.py`** - Core MPM algorithm implementation (copied from tuner-py)
- **`pitch_utils.py`** - Pitch to note conversion utilities
- **`test_mpm.py`** - Test suite for validating MPM accuracy
- **`testdata/`** - Test audio files (optional)

## Setup

```bash
# Create virtual environment
python3 -m venv venv
source venv/bin/activate

# Install dependencies
pip install -r requirements.txt
```

## Run Tests

```bash
python test_mpm.py
```

Expected output: ~0.06 cents average error on guitar-like tones.

## Usage as Reference

When porting to Go, use this as the reference:

1. Compare function signatures
2. Validate intermediate results (NSDF, peaks, etc.)
3. Match final detection accuracy

## Key Functions

### `mpm_pitch_detection(buffer, sample_rate, threshold)`
Main entry point - detects pitch for a single audio buffer.

**Returns:** `(frequency, clarity)`

### `normalized_square_difference(buffer)`
Core NSDF calculation using FFT-based autocorrelation.

**Returns:** `ndarray` of NSDF values

### `peak_picking(nsdf, threshold)`
Find positive peaks above threshold.

**Returns:** `list of (index, value)` tuples

### `parabolic_interpolation(nsdf, peak_index)`
Refine peak location for sub-sample accuracy.

**Returns:** `float` refined index

### `pitch_to_note(frequency)`
Convert Hz to musical note.

**Returns:** `(note_name, midi_number, cents_offset)`
