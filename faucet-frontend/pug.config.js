// Make .env available in pug
require('dotenv').config();
console.log('process.env.DOCUMENT_TITLE', process.env.DOCUMENT_TITLE);

module.exports = {
  pretty: true,
};
