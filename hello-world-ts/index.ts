import express from 'express';
import dotenv from 'dotenv';
import crypto from 'crypto';
import bodyParser from 'body-parser';

dotenv.config();

const app = express();
app.use(bodyParser.raw({
    type: 'application/json',
}));

const port = process.env.PORT || 3000;

app.get('/oauth/callback', (req, res) => {
    res.status(200).send(JSON.stringify({'ok': 'true'}));
});


app.post('/handle-turn', async (req, res) => {
    const signature = req.get("Github-Public-Key-Signature")!;
    const keyID = req.get("Github-Public-Key-Identifier")!;
    const tokenForUser = req.get("X-GitHub-Token")!;
    const payload = req.body.toString('utf8');

    try {
        await verifySignature(payload, signature, keyID, tokenForUser);
    } catch (e) {
        console.error(e);
        return res.status(401).send(JSON.stringify({ 'error': 'Invalid signature' }));
    }

    try {
        const user = (await fetch("https://api.github.com/user", {
            headers: {
                authorization: `token ${tokenForUser}`,
            },
        }).then((res) => res.json())) as { login: string };

        const body = JSON.parse(payload);
        let messages = body.messages;
        if (messages == null) {
            res.status(400).send(JSON.stringify({ 'error': 'Missing prompts in the request body' }));
            return;
        }

        messages = messages.map((m: any) => ({
            'role': m.role,
            'content': m.content
        }));
        messages.unshift({
            'role': 'system',
            'content': 'You are a cat! Think carefully and step by step like a cat would. Your job is to explain computer science concepts in the funny manner of a cat, using cat metaphors. Always start your response by stating what concept you are explaining. Always include code samples.'
        });

        const stream = await fetch(
            "https://api.githubcopilot.com/chat/completions",
            {
                method: "POST",
                headers: {
                    authorization: `Bearer ${tokenForUser}`,
                },
                body: JSON.stringify({
                    messages,
                    model: "gpt-3.5-turbo",
                    stream: true,
                }),
            }
        );

        res.contentType('text/event-stream');
        res.status(200);

        if (stream.body != null) {
            const reader = stream.body.getReader();
            let done, value;
            while (!done) {
                ({ value, done } = await reader.read());
                res.write(new TextDecoder().decode(value));
            }
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

interface GitHubKeysPayload {
  public_keys: Array<{
    key: string;
    key_identifier: string;
    is_current: boolean;
  }>;
}

const GITHUB_KEYS_URI = "https://api.github.com/meta/public_keys/copilot_api";

// verifySignature verifies the signature of a payload using the public key
// from GitHub's public key API. It fetches that public keys from GitHub's
// public key API, and uses the keyID to find the public key that signed the
// payload. It then verifies the signature using that public key.
async function verifySignature(
  payload: string,
  signature: string,
  keyID: string,
  tokenForUser: string
): Promise<void> {
  if (typeof payload !== "string" || payload.length === 0) {
    throw new Error("Invalid payload");
  }
  if (typeof signature !== "string" || signature.length === 0) {
    throw new Error("Invalid signature");
  }
  if (typeof keyID !== "string" || keyID.length === 0) {
    throw new Error("Invalid keyID");
  }

  const keys = (await fetch(GITHUB_KEYS_URI, {
    method: "GET",
    headers: {
      Authorization: `Bearer ${tokenForUser}`,
    },
  }).then((res) => res.json())) as GitHubKeysPayload;
  const publicKey = keys.public_keys.find((k) => k.key_identifier === keyID);
  if (!publicKey) {
    throw new Error("No public key found matching key identifier");
  }

  const verify = crypto.createVerify("SHA256").update(payload);
  if (!verify.verify(publicKey.key, signature, "base64")) {
    throw new Error("Signature does not match payload");
  }
}