import { Gemini } from "../../services/gemini.js";
import { writeFile } from "fs/promises";

const CreatePackageJson = async (code) => {
  const prompt = `The response should only contain a package.json file, without any additional content. and add type : module too. here is code ${code}`;
  const res = await Gemini(prompt);
  // res = res.substring(5, res.length - 3);
  //   console.log(res, "HERE");
  writeFile("./execute/js/package-code.json", res.substring(7, res.length - 3));
};

export default CreatePackageJson;
