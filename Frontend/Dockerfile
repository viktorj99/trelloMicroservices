# Koristimo Node.js sliku
FROM node:16-alpine

# Postavljamo radni direktorijum
WORKDIR /app

# Kopiramo package fajlove i instaliramo zavisnosti
COPY package*.json ./
RUN npm install

# Kopiramo ostatak aplikacije
COPY . .

# Izlažemo port na kojem Vite radi
EXPOSE 5173

# Pokrećemo Vite development server
CMD ["npm", "run", "dev", "--", "--host"]
