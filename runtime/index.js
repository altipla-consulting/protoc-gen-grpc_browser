
const url = require('url');
const unset = require('lodash/unset');

let fetchFn = global.fetch;
if (!fetchFn) {
  fetchFn = require('node-fetch');
}


class Caller {
  constructor({server = '', authorization = '', hook = null} = {}) {
    this.server = server;
    this.authorization = authorization;
    this.hook = hook;
  }

  send(method, binding, req, hasBody, path) {
    let endpoint = url.parse(this.server + binding);
    delete endpoint.search;

    path.split('/').forEach(segment => {
      if (segment.startsWith('{')) {
        unset(req, segment.substring(1, segment.length - 1));
      }
    });

    let opts = {
      method,
      headers: {
        'Content-Type': 'application/json',
      },
    };
    
    if (hasBody) {
      opts.body = JSON.stringify(req || {});
    } else {
      endpoint.query = req;
    }

    if (this.authorization) {
      opts.headers.Authorization = this.authorization;
    }

    return fetchFn(url.format(endpoint), opts)
      .then(response => {
        // TODO(ernesto): Leer los errores GRPC aquí.
        if (response.status !== 200) {
          let err = new Error(response.statusText);
          err.response = response;
          throw err;
        }

        if (this.hook) {
          this.hook(response);
        }

        return response.json();
      });
  }
}


module.exports = {
  Caller,
};
