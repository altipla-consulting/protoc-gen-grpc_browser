
const url = require('url');
const unset = require('lodash/unset');
const flat = require('flat');

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
      endpoint.query = req ? flat(req) : {};
    }

    if (this.authorization) {
      opts.headers.Authorization = this.authorization;
    }

    return fetchFn(url.format(endpoint), opts)
      .then(response => {
        if (response.status !== 200) {
          let code = response.headers.get('grpc-status');
          let message = response.headers.get('grpc-message');
          if (code) {
            throw new GrpcError(parseInt(code, 10), message);
          }

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


class GrpcError {
  constructor(code, message) {
    this.code = code;
    this.message = message;
  }
}


class StatusCode {
  // OK is returned on success.
  static get OK() { return 0; }

  // Canceled indicates the operation was canceled (typically by the caller).
  static get Canceled() { return 1; }

  // Unknown error. An example of where this error may be returned is
  // if a Status value received from another address space belongs to
  // an error-space that is not known in this address space. Also
  // errors raised by APIs that do not return enough error information
  // may be converted to this error.
  static get Unknown() { return 2; }

  // InvalidArgument indicates client specified an invalid argument.
  // Note that this differs from FailedPrecondition. It indicates arguments
  // that are problematic regardless of the state of the system
  // (e.g., a malformed file name).
  static get InvalidArgument() { return 3; }

  // DeadlineExceeded means operation expired before completion.
  // For operations that change the state of the system, this error may be
  // returned even if the operation has completed successfully. For
  // example, a successful response from a server could have been delayed
  // long enough for the deadline to expire.
  static get DeadlineExceeded() { return 4; }

  // NotFound means some requested entity (e.g., file or directory) was
  // not found.
  static get NotFound() { return 5; }

  // AlreadyExists means an attempt to create an entity failed because one
  // already exists.
  static get AlreadyExists() { return 6; }

  // PermissionDenied indicates the caller does not have permission to
  // execute the specified operation. It must not be used for rejections
  // caused by exhausting some resource (use ResourceExhausted
  // instead for those errors). It must not be
  // used if the caller cannot be identified (use Unauthenticated
  // instead for those errors).
  static get PermissionDenied() { return 7; }

  // ResourceExhausted indicates some resource has been exhausted, perhaps
  // a per-user quota, or perhaps the entire file system is out of space.
  static get ResourceExhausted() { return 8; }

  // FailedPrecondition indicates operation was rejected because the
  // system is not in a state required for the operation's execution.
  // For example, directory to be deleted may be non-empty, an rmdir
  // operation is applied to a non-directory, etc.
  //
  // A litmus test that may help a service implementor in deciding
  // between FailedPrecondition, Aborted, and Unavailable:
  //  (a) Use Unavailable if the client can retry just the failing call.
  //  (b) Use Aborted if the client should retry at a higher-level
  //      (e.g., restarting a read-modify-write sequence).
  //  (c) Use FailedPrecondition if the client should not retry until
  //      the system state has been explicitly fixed. E.g., if an "rmdir"
  //      fails because the directory is non-empty, FailedPrecondition
  //      should be returned since the client should not retry unless
  //      they have first fixed up the directory by deleting files from it.
  //  (d) Use FailedPrecondition if the client performs conditional
  //      REST Get/Update/Delete on a resource and the resource on the
  //      server does not match the condition. E.g., conflicting
  //      read-modify-write on the same resource.
  static get FailedPrecondition() { return 9; }

  // Aborted indicates the operation was aborted, typically due to a
  // concurrency issue like sequencer check failures, transaction aborts,
  // etc.
  //
  // See litmus test above for deciding between FailedPrecondition,
  // Aborted, and Unavailable.
  static get Aborted() { return 10; }

  // OutOfRange means operation was attempted past the valid range.
  // E.g., seeking or reading past end of file.
  //
  // Unlike InvalidArgument, this error indicates a problem that may
  // be fixed if the system state changes. For example, a 32-bit file
  // system will generate InvalidArgument if asked to read at an
  // offset that is not in the range [0,2^32-1], but it will generate
  // OutOfRange if asked to read from an offset past the current
  // file size.
  //
  // There is a fair bit of overlap between FailedPrecondition and
  // OutOfRange. We recommend using OutOfRange (the more specific
  // error) when it applies so that callers who are iterating through
  // a space can easily look for an OutOfRange error to detect when
  // they are done.
  static get OutOfRange() { return 11; }

  // Unimplemented indicates operation is not implemented or not
  // supported/enabled in this service.
  static get Unimplemented() { return 12; }

  // Internal errors. Means some invariants expected by underlying
  // system has been broken. If you see one of these errors,
  // something is very broken.
  static get Internal() { return 13; }

  // Unavailable indicates the service is currently unavailable.
  // This is a most likely a transient condition and may be corrected
  // by retrying with a backoff.
  //
  // See litmus test above for deciding between FailedPrecondition,
  // Aborted, and Unavailable.
  static get Unavailable() { return 14; }

  // DataLoss indicates unrecoverable data loss or corruption.
  static get DataLoss() { return 15; }

  // Unauthenticated indicates the request does not have valid
  // authentication credentials for the operation.
  static get Unauthenticated() { return 16; }
}


module.exports = {
  Caller,
  StatusCode,
};
