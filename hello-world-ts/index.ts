import express, { Response } from 'express';
import dotenv from 'dotenv';
import OpenAI from 'openai';
import { ChatCompletionChunk } from 'openai/resources';

dotenv.config();

const app = express();
app.use(express.json());

const port = process.env.PORT || 3000;
const model = process.env.OPENAI_MODEL_NAME as string;
const apiKey = process.env.OPENAI_API_KEY;
const apiVersion = process.env.OPENAI_API_VERSION;
const apiBaseURL = process.env.OPENAI_API_BASE;

if (!apiKey) {
  throw new Error('The AZURE_OPENAI_API_KEY environment variable is missing or empty.');
}

const openai = new OpenAI({
    apiKey: apiKey,
    baseURL: `${apiBaseURL}openai/deployments/${model}`,
    defaultQuery: { 'api-version': apiVersion },
    defaultHeaders: { 'api-key': apiKey }
});

app.get('/oauth/callback', (req, res) => {
    res.status(200).send(JSON.stringify({'ok': 'true'}));
});


app.post('/handle-turn', async (req, res) => {
    if (req.headers['content-type'] != 'application/json') {
        res.status(400).send(JSON.stringify({ 'error': 'Invalid content type' }));
        return;
    }

    try {
        const body = req.body;
        let messages = body.messages;
        if (messages == null) {
            res.status(400).send(JSON.stringify({ 'error': 'Missing prompts in the request body' }));
            return;
        }

        messages = messages.map((m: any) => ({
            'role': m.role,
            'content': m.content
        }));
        messages.push({
            'role': 'system',
            'content': 'You are a cat! Think carefully and step by step like a cat would. Your job is to explain computer science concepts in the funny manner of a cat, using cat metaphors. Always start your response by stating what concept you are explaining. Always include code samples.'
        });

        const dataStream = await openai.chat.completions.create({
            model: model,
            messages: messages,
            stream: true
        });

        res.contentType('text/event-stream');
        res.status(200);
        
        for await (const chunk of dataStream) {
            res.write(`data: ${JSON.stringify(chunk)}\n\n`);
        }
        
        res.end();
    } catch (e) {
        console.log(e);
        res.status(400).send(JSON.stringify({ 'error': e }));
    }
});

app.listen(port, () => {
    console.log(`Server is running on port ${port}`);
});
