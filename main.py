import pydirectinput
import json
import subprocess
import time
import numpy as np
from PIL import Image, ImageGrab

# --- Config ---
EXE_PATH = "./2048_ai.exe"
# Colors for the official 2048 site
COLOR_MAP = {
    (205, 193, 180): 0, (238, 228, 218): 2, (237, 224, 200): 4,
    (242, 177, 121): 8, (245, 149, 99): 16, (246, 124, 95): 32,
    (246, 94, 59): 64, (237, 207, 114): 128, (237, 204, 97): 256,
    (237, 200, 80): 512, (237, 197, 63): 1024, (237, 194, 46): 2048
}

def get_closest_value(rgb):
    min_dist = float('inf')
    best_val = 0
    for color, val in COLOR_MAP.items():
        dist = np.linalg.norm(np.array(rgb) - np.array(color))
        if dist < min_dist:
            min_dist = dist
            best_val = val
    return best_val if min_dist < 20 else 0

def capture_grid():
    # Use pyautogui just for finding the location, pydirectinput for keys
    import pyautogui 
    try:
        location = pyautogui.locateOnScreen('template.png', confidence=0.8)
        if not location: return None
        
        left, top, width, height = map(int, location)
        screenshot = ImageGrab.grab(bbox=(left, top, left+width, top+height))
        
        cell_w, cell_h = width / 4, height / 4
        grid = []
        for y in range(4):
            row = []
            for x in range(4):
                rgb = screenshot.getpixel((int(x*cell_w + cell_w/2), int(y*cell_h + cell_h/2)))
                row.append(get_closest_value(rgb))
            grid.append(row)
        return grid
    except:
        return None

def main():
    print("Autoplayer starting... Switch to your 2048 tab NOW!")
    time.sleep(3) 

    while True:
        grid = capture_grid()
        if not grid or not any(any(row) for row in grid):
            continue

        grid_json = json.dumps({"grid": grid})
        result = subprocess.run([EXE_PATH, "--grid", grid_json, "--depth", "5"], capture_output=True, text=True)

        try:
            decision = json.loads(result.stdout)
            move = decision['best_move'].lower()
            
            # Send hardware-level keypress
            pydirectinput.press(move)
            print(f"Executed: {move}")
        except:
            continue

if __name__ == "__main__":
    main()