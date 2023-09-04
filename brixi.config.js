module.exports = {
    outDir: "./brixi",
    important: true,
    output: "production",
    colors: {
        white: "#ffffff",
        black: "#000000",
        grey: {
            50: "#FAFAFA",
            100: "#F4F4F5",
            200: "#E4E4E7",
            300: "#D4D4D8",
            400: "#A1A1AA",
            500: "#71717A",
            600: "#52525B",
            700: "#3F3F46",
            800: "#27272A",
            900: "#18181B",
        },
        blue: {
            50: "#EFF6FF",
            100: "#DBEAFE",
            200: "#BFDBFE",
            300: "#93C5FD",
            400: "#60A5FA",
            500: "#3B82F6",
            600: "#2563EB",
            700: "#1D4ED8",
            800: "#1E40AF",
            900: "#1E3A8A",
        },
        green: {
            50: "#ECFDF5",
            100: "#D1FAE5",
            200: "#A7F3D0",
            300: "#6EE7B7",
            400: "#34D399",
            500: "#10B981",
            600: "#059669",
            700: "#047857",
            800: "#065F46",
            900: "#064E3B",
        },
        red: {
            50: "#FFF1F2",
            100: "#FFE4E6",
            200: "#FECDD3",
            300: "#FDA4AF",
            400: "#FB7185",
            500: "#F43F5E",
            600: "#E11D48",
            700: "#BE123C",
            800: "#9F1239",
            900: "#881337",
        },
        yellow: {
            50: "#FFFBEB",
            100: "#FEF3C7",
            200: "#FDE68A",
            300: "#FCD34D",
            400: "#FBBF24",
            500: "#F59E0B",
            600: "#D97706",
            700: "#B45309",
            800: "#92400E",
            900: "#78350F",
        },
    },
    borders: {
        units: "px",
        styles: ["solid"],
        widths: [1],
        radius: [0.25, 0.5, 1],
    },
    variables: {
        "focus-ring": "1px auto var(--primary-500)",
        "focus-ring-offset": "5px",
    },
    shadows: {
        colors: {
            grey: "var(--black-hsl)",
        },
        sizes: {
            sm: `
                0px 1px 2px hsl(var(--shadow-color) / 0.1)
            `,
            md: `
                0px 2px 2px hsl(var(--shadow-color) / 0.1),
                0px 4px 4px hsl(var(--shadow-color) / 0.1),
                0px 6px 6px hsl(var(--shadow-color) / 0.1)
            `,
            lg: `
                0px 2px 2px hsl(var(--shadow-color) / 0.1),
                0px 4px 4px hsl(var(--shadow-color) / 0.1),
                0px 8px 8px hsl(var(--shadow-color) / 0.1),
                0px 16px 16px hsl(var(--shadow-color) / 0.1),
                0px 32px 32px hsl(var(--shadow-color) / 0.1)
            `,
            xl: `
                0px 2px 2px hsl(var(--shadow-color) / 0.1),
                0px 4px 4px hsl(var(--shadow-color) / 0.1),
                0px 8px 8px hsl(var(--shadow-color) / 0.1),
                0px 16px 16px hsl(var(--shadow-color) / 0.1),
                0px 32px 32px hsl(var(--shadow-color) / 0.1),
                0px 48px 48px hsl(var(--shadow-color) / 0.1),
                0px 64px 64px hsl(var(--shadow-color) / 0.1)
            `,
        },
    },
};
