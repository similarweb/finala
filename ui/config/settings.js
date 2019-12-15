const development = {
    webserver_endpoint: "http://localhost:9090",
}
const production = {
    webserver_endpoint: "",
}

var configuration = {}
switch(process.env.NODE_ENV){
    case 'development':
            configuration = development
        break;
    case 'production':
        configuration = Object.assign( development, production)
        break;
  }

module.exports = configuration

