import {postEndpoint, getByIdEndpoint} from './base.js'

export const options = {
    stages: [
      { duration: '30s', target: 300 },
      { duration: '10s', target: 300 },
      { duration: '5s', target: 2000 },
      { duration: '10s', target: 2000 },
      { duration: '5s', target: 300 },
      { duration: '10s', target: 300 },
      { duration: '5s', target: 2000 },
      { duration: '10s', target: 2000 },
      { duration: '5s', target: 300 },
      { duration: '10s', target: 300 },
      { duration: '30s', target: 0 }, 
    ],
  };

  export default function () {
    postEndpoint();
    getByIdEndpoint();
  }