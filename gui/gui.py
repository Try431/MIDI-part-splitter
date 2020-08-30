#!/usr/bin/env python3
import tkinter as tk
from tkinter.filedialog import *
import subprocess
import os
from functools import partial
from re import match


class MIDIConvertGUI(object):
    def __init__(self, parent):
        self.root = parent
        self.label = tk.Label(text="MIDI File Conversion")
        self.label.pack()
        self.frame = tk.Frame(parent)
        self.frame.pack()
        self.files_to_convert_text_box = tk.Text(
            width=150
        )
        self.choose_file_btn = tk.Button(
            text="Add file for conversion",
            width=25,
            height=2,
            bg="#34A2FE",
        )
        self.choose_dir_btn = tk.Button(
            text="Add directory for conversion",
            width=25,
            height=2,
            bg="#34A2FE",
        )
        self.start_conversion_btn = tk.Button(
            text="Start conversion",
            width=25,
            height=2,
            bg="#57FF00", 
        )
        self.clear_all_btn = tk.Button(
            text="Clear all entries",
            width=15,
            height=2,
            bg="red"
        )
        self.log_box = tk.Text(
            width=100, 
            bg="black", 
            fg="white", 
            insertbackground="white"
        )
        self.window = tk.Tk()


if __name__ == "__main__":
    root = tk.Tk()
    gui = MIDIConvertGUI(root)
    gui.window.wm_withdraw()
    gui.choose_file_btn.pack()
    gui.choose_dir_btn.pack()
    gui.start_conversion_btn.pack()
    gui.files_to_convert_text_box.pack()
    gui.clear_all_btn.pack()
    root.mainloop()
    