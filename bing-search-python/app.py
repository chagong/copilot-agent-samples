import json
import os
import openai
import requests
from flask import Flask, request, jsonify, Response
from flask_cors import CORS
from datetime import datetime

app = Flask(__name__)
CORS(app)

openai.api_key = os.getenv("OPENAI_API_KEY")
openai.azure_endpoint = os.getenv("OPENAI_API_BASE")
openai.api_type = "azure"
openai.api_version = os.getenv("OPENAI_API_VERSION")
model_name = os.getenv("OPENAI_MODEL_NAME")

functions = [
    {
        "name": "get_bing_search_results",
        "description": "Get the search results from Bing search engine.",
        "parameters": {
            "type": "object",
            "properties": {
                "term": {
                    "type": "string",
                    "description": "The search term."
                }
            },
            "required": ["term"],
        }
    }
]

prompts = """
- Do your best to answer user's questions.
- Find out what user really wants to know and give exhaustive answer.
- Never say to user you don't know the answer, instead try to get answers from Bing search.
- Don't make assumptions about what values to plug into functions. Ask for clarification if a user request is ambiguous.
- You should always conclude the answer from Bing earch results.
- Conclude the answer from snippet text frist in Bing search results.
"""


def get_bing_search_results(search_request):
    search_term = search_request.get("term")
    subscription_key = os.getenv("BING_ACCESS_KEY")
    assert subscription_key
    search_url = 'https://api.bing.microsoft.com/v7.0/search'

    headers = {"Ocp-Apim-Subscription-Key": subscription_key}
    params = {"q": search_term, "textDecorations": True, "textFormat": "HTML"}
    response = requests.get(search_url, headers=headers, params=params)
    response.raise_for_status()
    data = response.json()
    values = data.get('webPages', {}).get('value', [])

    return values


def generate(data_stream):
    try:
        for chunk in data_stream:
            yield f"data: {chunk.json()}\n\n"
    finally:
        data_stream.close()


@app.route('/oauth/callback')
def oauth_callback():
    return jsonify({'ok': 'true'}), 200


@app.route('/handle-turn', methods=['POST'])
def handle_turn():
    # Check the content type
    if request.headers['Content-Type'] != 'application/json':
        return jsonify({'error': 'Invalid Content-Type'}), 400

    try:
        # Check the request body
        data = request.json
        messages = data.get('messages')
        if messages is None:
            return jsonify({'error': 'Missing prompts in the request body'}), 400

        messages = list(
            map(lambda x: {'role': x['role'], 'content': x['content']}, messages))
        messages.insert(
            0,
            {
                "role": "system",
                "content": prompts
            }
        )
        chat_completion = openai.chat.completions.create(
            model=model_name,
            messages=messages,
            functions=functions,
        )

        if chat_completion.choices[0].finish_reason == 'function_call':
            function_call = chat_completion.choices[0].message.function_call

            if function_call.name == 'get_bing_search_results':
                response = get_bing_search_results(
                    json.loads(function_call.arguments))
                print(response)

            messages.append(
                {
                    "role": "function",
                    "name": "get_bing_search_results",
                    "content": json.dumps(response)
                }
            )

            data_stream = openai.chat.completions.create(
                model=model_name,
                messages=messages,
                # functions=functions,
                stream=True,
            )

            return Response(generate(data_stream), mimetype='text/event-stream')
        else:
            chunk_dict = {
                'id': "chunk",
                'object': "chat.completion.chunk",
                'created': int(datetime.now().timestamp()),
                'model': model_name,
                'choices': [
                    {
                        'index': 0,
                        'delta': {
                            'content': chat_completion.choices[0].message.content
                        },
                        'finish_reason': chat_completion.choices[0].finish_reason
                    }
                ]
            }
            return f"data: {json.dumps(chunk_dict)}\n\n"

    except Exception as e:
        return jsonify({'error': f'{e}'}), 400


if __name__ == '__main__':
    app.run(debug=True)
