import { exec } from 'child_process';

exec('docker build -t code-execution-container .', (error, stdout, stderr) => {
    if (error) {
      console.error(`Error: ${error.message}`);
    }
    if (stderr) {
      console.error(`Error: ${stderr}`);
    }
    
    console.log(`Output: ${stdout}`);



    exec('docker run --rm code-execution-container', (error, stdout, stderr) => {
        if (error) {
          console.error(`Error: ${error.message}`);
        }
        if (stderr) {
          console.error(`Error: ${stderr}`);
        }
        
        console.log(`Output: ${stdout}`);
      });
  });