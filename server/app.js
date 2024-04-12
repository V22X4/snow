import express from 'express';
import {router as executeRouter} from './routes/run.js'


const app = express();
const port = 3000;

app.use(express.json());
app.use('/api/v1', executeRouter);

export { app };