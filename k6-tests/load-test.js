import {authenticate, postEndpoint, getByIdEndpoint} from './base.js'

export const options = {
    stages: [
      { duration: '30s', target: 100 },
      { duration: '20s', target: 100 },
      { duration: '10s', target: 0 }, 
    ],
  };

  export default function () {
    const bearerToken = authenticate();
    postEndpoint(bearerToken);
    getByIdEndpoint(bearerToken);
  }