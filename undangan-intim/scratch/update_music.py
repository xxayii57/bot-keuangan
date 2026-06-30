import os
import json
import random

music_list = [
    "https://docs.google.com/uc?export=download&id=1yFT3UT1KbfZsi6VbNQb2N6Aez09V1vzW",
    "https://docs.google.com/uc?export=download&id=1DkRtKuvgLigJ3gNc3i10i5cDpxL_AJsZ",
    "https://docs.google.com/uc?export=download&id=1qHbgOn3xW_9V-49lMA3l7myTQwWgk9QG",
    "https://docs.google.com/uc?export=download&id=14DnbI1tK4JydVA0DQh17iG_FEVoYh7r1",
    "https://docs.google.com/uc?export=download&id=1aPSMZXhI6fTco9b5LhcEQfz099YQ34iL",
    "https://docs.google.com/uc?export=download&id=1pnC3PcKdlc3mSLqKGr6cwj0K3l1H5SDy",
    "https://docs.google.com/uc?export=download&id=1FEFo5oy9mG5lzACEk6KA60QGJ7BFkX6U",
    "https://docs.google.com/uc?export=download&id=1MQ9ogjvL8pPhyYB1sP3rJdE2NI3YhZeX",
    "https://docs.google.com/uc?export=download&id=1UtFNhtqvRngF09bglwgyzYixuVjYmpnR",
    "https://docs.google.com/uc?export=download&id=1rlGPKCPItKRSWgVhvJYhjQhJgGJnSVxF"
]

data_dir = "/var/www/html/data"
for filename in os.listdir(data_dir):
    if filename.endswith(".json") and not filename.startswith("reseller_"):
        filepath = os.path.join(data_dir, filename)
        try:
            with open(filepath, 'r') as f:
                data = json.load(f)
            
            old_url = data.get("musicUrl", "")
            new_url = random.choice(music_list)
            data["musicUrl"] = new_url
            
            with open(filepath, 'w') as f:
                json.dump(data, f, indent=2)
            print(f"Updated {filename}: {old_url} -> {new_url}")
        except Exception as e:
            print(f"Failed to update {filename}: {e}")
