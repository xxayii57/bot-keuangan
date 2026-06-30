import os
import json

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
source_path = os.path.join(data_dir, "alya-fajar.json")

if not os.path.exists(source_path):
    print("Error: alya-fajar.json not found on STB!")
    exit(1)

with open(source_path, 'r') as f:
    template = json.load(f)

# We have themes v1 to v9
for i in range(1, 10):
    theme_slug = f"v{i}"
    dest_name = f"demo-{theme_slug}.json"
    dest_path = os.path.join(data_dir, dest_name)
    
    # Copy template and update fields
    demo_data = template.copy()
    demo_data["theme"] = theme_slug
    
    # Set music based on (i - 1) % 10
    music_idx = (i - 1) % 10
    demo_data["musicUrl"] = music_list[music_idx]
    
    # Custom names/titles to make it look unique for each demo
    demo_data["brideName"] = f"Alya (Demo {theme_slug.upper()})"
    demo_data["groomName"] = f"Fajar (Demo {theme_slug.upper()})"
    
    with open(dest_path, 'w') as f:
        json.dump(demo_data, f, indent=2)
    print(f"Generated {dest_name} with theme={theme_slug} and musicUrl={music_list[music_idx]}")
