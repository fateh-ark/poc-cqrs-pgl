import {postEndpoint, getByIdEndpoint} from './base.js'

export const options = {
    stages: [
      { duration: '30s', target: 1000 },
      { duration: '1m', target: 1000 },
      { duration: '30s', target: 0 }, 
    ],
  };

  export default function () {
    postEndpoint();
    getByIdEndpoint();
  }