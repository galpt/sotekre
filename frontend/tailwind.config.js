module.exports = {
    content: [
        './pages/**/*.{js,ts,jsx,tsx}',
        './components/**/*.{js,ts,jsx,tsx}'
    ],
    theme: {
        extend: {
            colors: {
                primary: {
                    DEFAULT: '#0D47A1',
                    dark: '#083A89',
                    light: '#1565C0',
                }
            }
        }
    },
    plugins: []
}
