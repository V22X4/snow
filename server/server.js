import express from 'express';
import { writeFile } from 'fs/promises';
import { exec } from 'child_process';


const app = express();
const port = 3000;

app.use(express.json());

app.post('/re', (req, res) => {
});

app.post('/', (req, res) => {
    const { code } = req.body;
    

  if (!code) {
    return res.status(400).send('No code provided');
  }

  try {

    writeFile('./js/code.js', code);

    exec('docker build -t code-execution-container -f ./js/Dockerfile .', (error, stdout, stderr) => {
      if (error) {
        console.error(`Error: ${error.message}`);
        return res.status(500).send('Error executing code hh');
      }
        
    
      exec('docker run --rm code-execution-container', (error, stdout, stderr) => {
        if (error) {
          console.error(`Docker run error: ${error.message}`);
          return res.status(500).send('Error running Docker container');
        }
        else if (stderr) {
          console.error(`Docker run stderr: ${stderr}`);
          return res.status(500).send('Error running Docker container');
        }
  
        
        res.send(stdout);
      });
    });

  } catch (error) {
    res.status(500).send('Error executing code');
  }
});

app.listen(port, () => {
  console.log(`Server is running on http://localhost:${port}`);
});