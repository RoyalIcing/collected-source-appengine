const Path = require('path')

module.exports = {
  plugins: [
    require('postcss-import'),
    require('tailwindcss')(Path.join(__dirname, 'tailwind.js')),
    require('autoprefixer'),
  ]
}