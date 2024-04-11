import pg from "pg";
import dotenv from "dotenv";

dotenv.config();

const connectionString = `postgresql://${process.env.POSTGRES_USER}:${process.env.POSTGRES_PASSWORD}@${process.env.POSTGRES_HOST}:${process.env.POSTGRES_PORT}/${process.env.database}`;

//const connectionString = 'postgresql://postgres:postgres@localhost:5432/testdatabase';
console.log(connectionString)

const pool = new pg.Pool({
  connectionString,
});


const query = (text, params) => pool.query(text, params);
const end = () => pool.end();

export { query, end };
