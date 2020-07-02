const PROXY_CONFIG = {
  '/registerLink': {
    'bypass': (req, res) => {
      const result = {
        shortenedLink: "http://localhost:4200/trololo"
      };

      res.end(JSON.stringify(result));
      return true;
    }
  }
}

module.exports = PROXY_CONFIG;
