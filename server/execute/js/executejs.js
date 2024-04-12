import { writeFile } from 'fs/promises';
import { exec } from 'child_process';
import CreatePackageJson from './CreatePackageJson.js';


const executejs = async (req, res) => {


  const { code } = req.body;



    if (!code) {
        return res.status(400).send('No code provided');
    }
    
  try {
      
      const absPath = './execute/js'
      await Promise.all([writeFile(`${absPath}/code.js`, code), CreatePackageJson(code)]);
      
    
      await exec(`docker build -t js-code-execution-container -f ${absPath}/Dockerfile .`, (error, stdout, stderr) => {
          if (error) {
            console.error(`Error: ${error.message}`);
            return res.status(500).send('Error executing code hh');
          }
        });
        
        
      await exec('docker run --rm js-code-execution-container', (error, stdout, stderr) => {
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
    
    } catch (error) {
        res.status(500).send('Error executing code');
    }
}

export { executejs };