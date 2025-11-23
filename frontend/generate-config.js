import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));

let mode = process.env.MODE === 'prod' ? 'prod' : 'dev';

const configs = {
  dev: {
    "API_URL": "/hack",
    "API_UPLOAD_URL": "/hack",
    "GOOGLE_CLIENT_ID": "fake_client_id",
    "GOOGLE_OAUTH_URL": "/hack/fakeoauth/oauth2/v2/auth",
    "GOOGLE_REDIRECT_URL": "/hack/callback"
  },
  prod: {
    "API_URL": "https://internal.sortedchat.com",
    "API_UPLOAD_URL": "https://internal.sortedchat.com",
    "GOOGLE_CLIENT_ID": "120883379229-eof7t025att2p5heoiueng8crnifaosl.apps.googleusercontent.com",
    "GOOGLE_OAUTH_URL": "http://accounts.google.com/o/oauth2/v2/auth",
    "GOOGLE_REDIRECT_URL": "https://internal.sortedchat.com/callback"
  }
};

const config = configs[mode];

if (!config) {
  throw new Error(`Unknown MODE: ${mode}`);
}

const outputPath = path.join(__dirname, '../frontend/public/ui-config.json');
fs.writeFileSync(outputPath, JSON.stringify(config, null, 2));
console.log(`ui-config.json generated for mode: ${mode}`);
