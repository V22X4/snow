import { writeFile } from 'fs/promises';
import { spawn } from 'child_process';
import CreatePackageJson from './CreatePackageJson.js';

const executejs = async (req, res) => {
    const { code } = req.body;

    if (!code) {
        return res.status(400).send('No code provided');
    }

    try {
        const absPath = './execute/js';
        await Promise.all([writeFile(`${absPath}/code.js`, code), CreatePackageJson(code)]);

        // Build Docker image
        const buildProcess = spawn('docker', ['build', '-t', 'js-code-execution-container', '-f', `${absPath}/Dockerfile`, '.']);

        buildProcess.stderr.on('data', (data) => {
            console.error(`Error during image build: ${data}`);
        });

        buildProcess.on('close', async (code) => {
            if (code !== 0) {
                console.error(`Docker build process exited with code ${code}`);
                return res.status(500).send('Error executing code');
            }

            // Run Docker container
            const runProcess = spawn('docker', ['run', '--rm', 'js-code-execution-container']);

            let stdout = '';
            let stderr = '';

            runProcess.stdout.on('data', (data) => {
                stdout += data;
            });

            runProcess.stderr.on('data', (data) => {
                stderr += data;
            });

            runProcess.on('close', (code) => {
                if (code !== 0) {
                    console.error(`Docker run process exited with code ${code}`);
                    return res.status(500).send('Error executing code');
                }

                if (stderr) {
                    console.error(`Docker run stderr: ${stderr}`);
                    return res.status(200).send(stderr);
                }

                res.send(stdout);
            });
        });
    } catch (error) {
        console.error(`Error executing code: ${error}`);
        res.status(500).send('Error executing code');
    }
};

export { executejs };
