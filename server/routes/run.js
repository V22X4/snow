import express from "express";
import { executejs } from "../execute/execute.js";

const router = express.Router();

router.route('/run/js').post(executejs);

export { router };