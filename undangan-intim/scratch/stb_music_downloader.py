
import os
import json
import urllib.request
import ssl

ssl._create_default_https_context = ssl._create_unverified_context

music_dir = "/var/www/html/assets/music"
os.makedirs(music_dir, exist_ok=True)

files = [{'id': '1yFT3UT1KbfZsi6VbNQb2N6Aez09V1vzW', 'name': 'sampai_tua_nanti.mp3'}, {'id': '1DkRtKuvgLigJ3gNc3i10i5cDpxL_AJsZ', 'name': 'aku_memilihmu.mp3'}, {'id': '1qHbgOn3xW_9V-49lMA3l7myTQwWgk9QG', 'name': 'cinta_terakhir.mp3'}, {'id': '14DnbI1tK4JydVA0DQh17iG_FEVoYh7r1', 'name': 'on_this_day.mp3'}, {'id': '1aPSMZXhI6fTco9b5LhcEQfz099YQ34iL', 'name': 'bersamamu.mp3'}, {'id': '1pnC3PcKdlc3mSLqKGr6cwj0K3l1H5SDy', 'name': 'ketika_cinta_bertasbih.mp3'}, {'id': '1FEFo5oy9mG5lzACEk6KA60QGJ7BFkX6U', 'name': 'for_the_rest_of_my_life.mp3'}, {'id': '1MQ9ogjvL8pPhyYB1sP3rJdE2NI3YhZeX', 'name': 'penjaga_hati.mp3'}, {'id': '1UtFNhtqvRngF09bglwgyzYixuVjYmpnR', 'name': 'akad.mp3'}, {'id': '1rlGPKCPItKRSWgVhvJYhjQhJgGJnSVxF', 'name': 'nahalal_kawin.mp3'}]

print("Downloading music files directly to STB assets/music/...")
for f in files:
    url = f"https://drive.google.com/uc?export=download&id={f['id']}"
    path = os.path.join(music_dir, f['name'])
    
    # Check if file already exists with non-zero size
    if os.path.exists(path) and os.path.getsize(path) > 100000:
        print(f"Skipping {f['name']} (already downloaded)")
        continue
        
    try:
        req = urllib.request.Request(
            url,
            headers={'User-Agent': 'Mozilla/5.0'}
        )
        with urllib.request.urlopen(req) as response:
            with open(path, 'wb') as out_file:
                out_file.write(response.read())
        print(f"Successfully downloaded {f['name']}")
    except Exception as e:
        print(f"Failed to download {f['name']}: {e}")

# Update JSON files
data_dir = "/var/www/html/data"
music_names = [f['name'] for f in files]

print("Updating client JSON data files to use local STB music paths...")
for filename in os.listdir(data_dir):
    if filename.endswith(".json") and not filename.startswith("reseller_"):
        filepath = os.path.join(data_dir, filename)
        try:
            with open(filepath, 'r') as json_file:
                data = json.load(json_file)
            
            selected_music = files[hash(filename) % len(files)]['name']
            new_url = f"https://intim.my.id/assets/music/{selected_music}"
            
            data["musicUrl"] = new_url
            with open(filepath, 'w') as json_file:
                json.dump(data, json_file, indent=2)
            print(f"Updated {filename} -> {new_url}")
        except Exception as e:
            pass
