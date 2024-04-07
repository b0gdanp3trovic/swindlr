const express = require('express');
const app = express();
const port = 3000;

app.get('/', (req, res) => {
  const containerId = process.env.CONTAINER_ID || 'Unknown';
  console.log(`Sending container ID: ${containerId}`);
  res.send(`Container ID: ${containerId}`);
});

app.listen(port, () => {
  console.log(`Application listening at http://localhost:${port}`);
});