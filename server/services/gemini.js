const Gemini = async (message) => {

    const url = `https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent?key=` + process.env.GEMINI_API_KEY

    // console.log(message)

    const messagesToSend = [
        // {
        //     "role": "user",
        //     "parts": [{
        //         "text": prompt
        //     }],
        // },
        // {
        //     "role": "model",
        //     "parts": [{
        //         "text": "sure, I will help you with that"
        //     }],
        // },
        {
            "role": "user",
            "parts": [{
                "text" : message,
            }],
        }
    ]


    const res = await fetch(url, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            "contents": messagesToSend
        })
    })


    const resjson = await res.json()

    const responseMessage = resjson.candidates[0].content.parts[0].text

    // console.log(responseMessage, "here")

    return responseMessage;

}

export { Gemini };