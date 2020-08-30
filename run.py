#!/usr/bin/env python3
import tkinter as tk
from tkinter.filedialog import *
import subprocess
import os
from functools import partial


# https://docs.python.org/3/library/subprocess.html


def remove_duplicates(text_box):
    line_set = set()
    for line in text_box.get('1.0', 'end-1c').splitlines():
        line_set.add(line)
    i = 1
    text_box.config(state="normal")
    text_box.delete("1.0", END)
    for line in line_set:
        text_box.insert(f"{i}.0", line+"\n")
        i+=1
    text_box.config(state="disabled")
    return text_box

def add_file_to_text_box(text_box, paths):
    for path in paths:
        if os.path.isdir(path):
            files_in_dir = map(lambda p: os.path.join(path, p), os.listdir(path))
            add_file_to_text_box(text_box, files_in_dir)
        elif os.path.isfile(path):
            text_box.config(state="normal")
            text_box.insert(tk.END, path+"\n")
            remove_duplicates(text_box)
            text_box.config(state="disabled")

def select_file(text_box, event):
    filename = askopenfilename()
    if filename:
        add_file_to_text_box(text_box, [filename])

def select_dir(text_box, event):
    directory = askdirectory()
    if directory:
        files_in_dir = map(lambda p: os.path.join(directory, p), os.listdir(directory))
        add_file_to_text_box(text_box, files_in_dir)

def main():
    window = tk.Tk()
    greeting = tk.Label(text="MIDI File Conversion")
    greeting.pack()
    
    choose_file_btn = tk.Button(
        text="Add file for conversion",
        width=25,
        height=2,
        bg="#34A2FE",
    )
    choose_dir_btn = tk.Button(
        text="Add directory for conversion",
        width=25,
        height=2,
        bg="#34A2FE",
    )
    start_conversion_btn = tk.Button(
        text="Start conversion",
        width=25,
        height=2,
        bg="#57FF00", 
    )
    
    text_box = tk.Text(width=150)
    choose_file_btn.pack()
    choose_dir_btn.pack()
    choose_file_btn.bind("<Button-1>", partial(select_file, text_box))
    choose_dir_btn.bind("<Button-1>", partial(select_dir, text_box))
    start_conversion_btn.pack()
    text_box.config(state="disabled")
    text_box.pack()
    window.mainloop()
    
main()

# pp = pprint.PrettyPrinter(indent=2)
# Tk().withdraw() # we don't want a full GUI, so keep the root window from appearing

# a = subprocess.run(['./MIDI-part-splitter', '-f', filename], capture_output=True)
# output = a.stdout
# args = a.args

# output_split = output.decode().split('\n')
# print()
# print(output_split)
# print(args)