# MyFirstCopilotPlugin
A simple Copilot plugin that answer user's questions using Bing search service.

# Debugging
- create ".env" file under workspace root folder.
- Fill in the following env variables like below example:
```
    OPENAI_API_BASE=https://xxxx/
    OPENAI_API_TYPE=azure
    OPENAI_API_VERSION=2023-07-01-preview
    OPENAI_MODEL_NAME=gpt-4
    OPENAI_API_KEY=<Your OpenAI KEY>
    BING_ACCESS_KEY=<Bing service access key>
```
- Start debug or press F5
