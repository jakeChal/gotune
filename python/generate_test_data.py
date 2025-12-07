#!/usr/bin/env python3
"""
Generate golden test data for validating the Go MPM implementation.
This creates JSON files with input signals and expected outputs.
"""

import json
import numpy as np
from mpm import normalized_square_difference


def generate_sine_wave(frequency, duration, sample_rate):
    """Generate a sine wave signal."""
    t = np.linspace(0, duration, int(sample_rate * duration), endpoint=False)
    return np.sin(2 * np.pi * frequency * t)


def main():
    test_cases = []

    # Test case 1: Simple 440Hz sine wave
    signal_440 = generate_sine_wave(440, 0.1, 48000)
    nsdf_440 = normalized_square_difference(signal_440)

    test_cases.append({
        "name": "440Hz sine wave",
        "input": signal_440.tolist(),
        "expected_nsdf": nsdf_440.tolist(),
        "sample_rate": 48000,
        "frequency": 440
    })

    # Test case 2: Lower frequency 220Hz
    signal_220 = generate_sine_wave(220, 0.1, 48000)
    nsdf_220 = normalized_square_difference(signal_220)

    test_cases.append({
        "name": "220Hz sine wave",
        "input": signal_220.tolist(),
        "expected_nsdf": nsdf_220.tolist(),
        "sample_rate": 48000,
        "frequency": 220
    })

    # Test case 3: Higher frequency 880Hz
    signal_880 = generate_sine_wave(880, 0.1, 48000)
    nsdf_880 = normalized_square_difference(signal_880)

    test_cases.append({
        "name": "880Hz sine wave",
        "input": signal_880.tolist(),
        "expected_nsdf": nsdf_880.tolist(),
        "sample_rate": 48000,
        "frequency": 880
    })

    # Test case 4: Small buffer
    signal_small = generate_sine_wave(440, 0.01, 48000)
    nsdf_small = normalized_square_difference(signal_small)

    test_cases.append({
        "name": "small buffer",
        "input": signal_small.tolist(),
        "expected_nsdf": nsdf_small.tolist(),
        "sample_rate": 48000,
        "frequency": 440
    })

    # Save to JSON
    output = {
        "version": "1.0",
        "description": "Golden test data for NSDF validation (via generate_test_data.py)",
        "test_cases": test_cases
    }

    with open("testdata/nsdf_golden.json", "w") as f:
        json.dump(output, f, indent=2)

    print(f"Generated {len(test_cases)} test cases")
    print("Saved to testdata/nsdf_golden.json")


if __name__ == "__main__":
    main()
