const http = require("http");

const server = http.createServer((request, response) => {
  switch (request.url) {
    case "/process":
      response.end(JSON.stringify(process.env));
      break;

    case "/node/version":
      response.end(process.version);
      break;

    default:
      response.end("Hello World!");
  }
});

const port = process.env.PORT || 8080;
server.listen(port, (err) => {
  if (err) {
    return console.log("something bad happened", err);
  }

  console.log(`NOT vendored server is listening on ${port}`);
});
