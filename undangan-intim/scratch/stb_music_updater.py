import os
import json
import re

music_list = [
    "https://intim.my.id/assets/music/sampai_tua_nanti.mp3",
    "https://intim.my.id/assets/music/aku_memilihmu.mp3",
    "https://intim.my.id/assets/music/cinta_terakhir.mp3",
    "https://intim.my.id/assets/music/on_this_day.mp3",
    "https://intim.my.id/assets/music/bersamamu.mp3",
    "https://intim.my.id/assets/music/ketika_cinta_bertasbih.mp3",
    "https://intim.my.id/assets/music/for_the_rest_of_my_life.mp3",
    "https://intim.my.id/assets/music/penjaga_hati.mp3",
    "https://intim.my.id/assets/music/akad.mp3",
    "https://intim.my.id/assets/music/nahalal_kawin.mp3"
]

data_dir = "/var/www/html/data"
for filename in os.listdir(data_dir):
    if filename.endswith(".json") and not filename.startswith("reseller_") and not filename.endswith("_comments.json") and not filename.endswith("-wishes.json"):
        filepath = os.path.join(data_dir, filename)
        try:
            with open(filepath, 'r') as f:
                data = json.load(f)
            
            theme = data.get("theme", "v1")
            match = re.search(r'\d+', theme)
            num = int(match.group(0)) if match else 1
            idx = (num - 1) % 10
            
            new_url = music_list[idx]
            data["musicUrl"] = new_url
            
            with open(filepath, 'w') as f:
                json.dump(data, f, indent=2)
            print(f"Updated {filename} (Theme {theme}) -> {new_url}")
        except Exception as e:
            print(f"Failed {filename}: {e}")
