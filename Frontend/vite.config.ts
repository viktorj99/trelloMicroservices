import { defineConfig } from 'vite';
import fs from 'fs';
import path from 'path';

export default defineConfig({
	server: {
		https: {
			key: fs.readFileSync(path.resolve(__dirname, 'certificates/cert.key')),
			cert: fs.readFileSync(path.resolve(__dirname, 'certificates/cert.crt')),
		},
		watch: {
			usePolling: true,
		},
	},
});
