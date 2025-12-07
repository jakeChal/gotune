"""Pitch conversion utilities - minimal version for MPM testing."""

import numpy as np


def pitch_to_note(frequency):
    """
    Convert a frequency in Hz to the nearest musical note.

    Parameters
    ----------
    frequency : float
        Frequency in Hz

    Returns
    -------
    note_name : str
        Musical note name (e.g., 'A4', 'C#3')
    midi_number : int
        MIDI note number
    cents_offset : float
        Offset in cents from the nearest note
    """
    if frequency is None or frequency <= 0:
        return None, None, None

    # A4 = 440 Hz = MIDI note 69
    midi_number = 69 + 12 * np.log2(frequency / 440.0)
    midi_number_rounded = int(round(midi_number))

    # Calculate cents offset from nearest note
    cents_offset = 100 * (midi_number - midi_number_rounded)

    # Convert MIDI number to note name
    note_names = ['C', 'C#', 'D', 'D#', 'E', 'F',
                  'F#', 'G', 'G#', 'A', 'A#', 'B']
    octave = (midi_number_rounded // 12) - 1
    note = note_names[midi_number_rounded % 12]
    note_name = f"{note}{octave}"

    return note_name, midi_number_rounded, cents_offset
