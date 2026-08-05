const express = require('express');
const morgan = require('morgan')
const path = require('path');
const app = express();
const port = process.env.PORT || 8080;

app.set('view engine', 'ejs');
app.set('views', path.join(__dirname, 'views'));

app.use(morgan('short', {
    skip: (req, res) => req.url === '/health' || req.url === '/healthcheck'
}));

const quotes = [
    "Код работает? Не трогай.",
    "Семь раз отмерь, один раз запушь.",
    "В любой непонятной ситуации делай git status."
];

app.get('/', (req, res) => {
    res.render('index');
});

app.get('/quote', (req, res) => {
    const randomIndex = Math.floor(Math.random() * quotes.length);
    res.render('quote', { quote: quotes[randomIndex] });
});

app.get('/health', (req, res) => {
    res.json({ status: "ok" });
});

app.listen(port, () => {
    console.log(`[${new Date().toISOString()}] INFO: Server started on port ${port}`);
});
