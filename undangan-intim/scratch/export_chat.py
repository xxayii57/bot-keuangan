import json
import os

transcript_path = "/data/data/com.termux/files/home/.gemini/antigravity-cli/brain/6840c84b-fc07-470b-bc4c-ef51171f985b/.system_generated/logs/transcript.jsonl"
output_md_path = "/data/data/com.termux/files/home/undangan-intim/chat_history.md"
output_jsonl_path = "/data/data/com.termux/files/home/undangan-intim/chat_log.jsonl"

# Copy raw jsonl
if os.path.exists(transcript_path):
    try:
        with open(transcript_path, "r", encoding="utf-8") as f:
            lines = f.readlines()
        with open(output_jsonl_path, "w", encoding="utf-8") as f_out:
            f_out.writelines(lines)
        print("Copied raw JSONL log successfully.")
    except Exception as e:
        print("Error copying JSONL:", e)

# Parse and create beautiful Markdown chat history
if os.path.exists(transcript_path):
    try:
        markdown_content = "# Antigravity Chat History\n\nThis file contains the complete chronological conversation history for the wedding invitation project.\n\n---\n\n"
        
        with open(transcript_path, "r", encoding="utf-8") as f:
            for line in f:
                if not line.strip():
                    continue
                step = json.loads(line)
                
                # Check for user input
                if step.get("type") == "USER_INPUT" and step.get("source") == "USER_EXPLICIT":
                    content = step.get("content", "")
                    # Strip checkpoint notices
                    if "{{ CHECKPOINT" in content:
                        continue
                    markdown_content += f"### 👤 USER\n\n{content}\n\n---\n\n"
                
                # Check for model responses
                elif step.get("type") == "PLANNER_RESPONSE" and step.get("source") == "MODEL":
                    content = step.get("content", "")
                    if content.strip():
                        markdown_content += f"### 🤖 ANTIGRAVITY / SUPERAGENT\n\n{content}\n\n---\n\n"
                        
        with open(output_md_path, "w", encoding="utf-8") as f_out:
            f_out.write(markdown_content)
        print("Exported chat history to markdown successfully.")
    except Exception as e:
        print("Error exporting markdown:", e)
else:
    print("Transcript log file not found.")
