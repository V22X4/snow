import { app } from './app.js'
import dotenv from 'dotenv'

dotenv.config();

// const connectDB = require('./config/database');

// connectDB();

const server = app.listen(process.env.PORT , ()=>{
    console.log(`listening on port ${process.env.PORT}`);
})

// Uncaught Exception 
process.on("uncaughtException" , (err)=> {
    console.log(err.message);
    console.log("server shutting down" );
    server.close(()=>{
        process.exit(1);
    });
})

// Unhandled Promise errors
process.on("unhandledRejection" , (err)=> {
    console.log(err.message);
    console.log("server shutting down" );
    server.close(()=>{
        process.exit(1);
    });
})