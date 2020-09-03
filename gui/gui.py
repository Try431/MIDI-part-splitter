#!/usr/bin/env python3
import tkinter as tk
import tkinter.filedialog as tkfd
import subprocess
import os
from functools import partial
import re

"""
TODO:
- Add confirmation dialog for clear_all button
- Create separate frame for clear_all button, so that it's harder to click accidentally
- Associate functions with buttons (referencing class attributes instead of parameters)
- 
"""
class MIDIConvertGUI(object):
    def __init__(self, parent):
        self.root = parent
        self.label = tk.Label(text="MIDI File Conversion")
        self.label.pack()
        self.button_frame = tk.Frame(parent)
        self.button_frame.pack(side=tk.TOP)
        self.files_to_convert_text_box = tk.Text(
            width=150
        )
        self.choose_file_btn = tk.Button(
            self.button_frame,
            text="Add file for conversion",
            width=25,
            height=2,
            bg="#34A2FE",
        )
        self.choose_dir_btn = tk.Button(
            self.button_frame,
            text="Add directory for conversion",
            width=25,
            height=2,
            bg="#34A2FE",
        )
        self.start_conversion_btn = tk.Button(
            self.button_frame,
            text="Start conversion",
            width=25,
            height=2,
            bg="#57FF00", 
        )
        self.clear_all_btn = tk.Button(
            self.button_frame,
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
        
        
    def select_file(self, event):
        filename = tkfd.askopenfilename()
        if filename:
            self.add_file_to_text_box([filename])
            
    def select_dir(self, event):
        directory = tkfd.askdirectory()
        if directory:
            files_in_dir = map(lambda p: os.path.join(directory, p), os.listdir(directory))
            self.add_file_to_text_box(files_in_dir)
            
            
    def remove_duplicates(self):
        line_set = set()
        current_lines = self.files_to_convert_text_box.get('1.0', 'end-1c').splitlines()
        for line in current_lines:
            line_set.add(line)
            
        if len(current_lines) == len(line_set):
            return
        
        i = 1
        self.files_to_convert_text_box.config(state="normal")
        self.files_to_convert_text_box.delete("1.0", tk.END)
        for line in line_set:
            self.files_to_convert_text_box.insert(f"{i}.0", line+"\n")
            i+=1
        self.files_to_convert_text_box.config(state="disabled")

    def add_file_to_text_box(self, paths):
        for path in paths:
            if os.path.isdir(path):
                files_in_dir = map(lambda p: os.path.join(path, p), os.listdir(path))
                self.add_file_to_text_box(files_in_dir)
            elif os.path.isfile(path):
                full_path, filename = os.path.split(path)
                if not re.match("mid*", filename.partition(".")[2]):
                    self.log_box.insert(tk.END, f"{path} does not point to a MIDI file\n")
                else:
                    self.files_to_convert_text_box.config(state="normal")
                    self.files_to_convert_text_box.insert(tk.END, path+"\n")
                    self.remove_duplicates()
                    self.files_to_convert_text_box.config(state="disabled")

        
            

            
    # def begin_conversion(text_box, log_box, event):
    #     lines_to_convert = text_box.get('1.0', 'end-1c').splitlines()
    #     # for line in lines_to_convert:


if __name__ == "__main__":
    root = tk.Tk()
    gui = MIDIConvertGUI(root)
    gui.choose_file_btn.pack(side=tk.LEFT)
    gui.choose_dir_btn.pack(side=tk.LEFT)
    gui.start_conversion_btn.pack(side=tk.RIGHT)
    gui.clear_all_btn.pack(side=tk.BOTTOM)
    gui.files_to_convert_text_box.pack()
    gui.log_box.pack()
    
    gui.choose_file_btn.bind("<Button-1>", gui.select_file)
    gui.choose_dir_btn.bind("<Button-1>", gui.select_dir)
    
    root.mainloop()
    
    # Old code to convert
    
    # files_to_convert_text_box = tk.Text(width=150)
    # log_box = tk.Text(width=100, bg="black", fg="white", insertbackground="white")
    # choose_file_btn.pack()
    # choose_dir_btn.pack()
    # clear_all_btn.pack()
    # choose_file_btn.bind("<Button-1>", partial(select_file, files_to_convert_text_box, log_box))
    # choose_dir_btn.bind("<Button-1>", partial(select_dir, files_to_convert_text_box, log_box))
    # start_conversion_btn.bind("<Button-1>", partial(begin_conversion, files_to_convert_text_box))
    # start_conversion_btn.pack()
    # files_to_convert_text_box.config(state="disabled")
    # files_to_convert_text_box.pack()
    # log_box.pack()
    # window.mainloop()
    
# pp = pprint.PrettyPrinter(indent=2)
# Tk().withdraw() # we don't want a full GUI, so keep the root window from appearing

# a = subprocess.run(['./MIDI-part-splitter', '-f', filename], capture_output=True)
# output = a.stdout
# args = a.args

# output_split = output.decode().split('\n')
# print()
# print(output_split)
# print(args)