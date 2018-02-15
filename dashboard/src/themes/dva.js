import createTheme from "./createTheme";
import colors from "../colors";

const shadowKeyUmbraOpacity = 0.2;
const shadowKeyPenumbraOpacity = 0.14;
const shadowAmbientShadowOpacity = 0.12;

function createShadow(...px) {
  return [
    `${px[0]}px ${px[1]}px ${px[2]}px ${
      px[3]
    }px rgba(0, 0, 0, ${shadowKeyUmbraOpacity})`,
    `${px[4]}px ${px[5]}px ${px[6]}px ${
      px[7]
    }px rgba(0, 0, 0, ${shadowKeyPenumbraOpacity})`,
    `${px[8]}px ${px[9]}px ${px[10]}px ${
      px[11]
    }px rgba(0, 0, 0, ${shadowAmbientShadowOpacity})`,
  ].join(",");
}

const dva = (type = "dark") =>
  createTheme({
    palette:
      type === "dark"
        ? {
            type,

            // secondary: {
            //   light: colors.cyberGrape[300],
            //   main: colors.cyberGrape[500],
            //   dark: colors.cyberGrape[600],
            //   contrastText: "#F1F9FD",
            // },
            secondary: {
              light: "#01EC8F",
              main: "#01EC8F",
              dark: "#01EC8F",
              contrastText: "#F1F9FD",
            },
            primary: {
              light: colors.persianPink[300],
              main: colors.persianPink[500],
              dark: colors.persianPink[700],
              contrastText: "#F1F9FD",
            },

            background: {
              default: colors.cyberGrape[500],
            },
          }
        : {
            type,

            primary: {
              light: colors.cyberGrape[300],
              main: colors.cyberGrape[500],
              dark: colors.cyberGrape[600],
            },
            secondary: {
              light: "#01EC8F",
              main: "#01EC8F",
              dark: "#01EC8F",
            },

            background: {
              paper: "#F1F9FD",
            },
          },

    shadows: [
      "none",
      createShadow(0, 1, 1, 0, 0, 1, 1, 0, 0, 1, 1, -1),
      createShadow(0, 1, 1, 0, 0, 1, 1, 0, 0, 1, 1, -1),
      createShadow(0, 1, 1, 0, 0, 1, 1, 0, 0, 1, 1, -1),
      createShadow(0, 1, 1, 0, 0, 1, 1, 0, 0, 1, 1, -1),
      createShadow(0, 1, 1, 0, 0, 1, 1, 0, 0, 1, 1, -1),
      createShadow(0, 1, 1, 0, 0, 1, 1, 0, 0, 1, 1, -1),
      createShadow(0, 1, 1, 0, 0, 1, 1, 0, 0, 1, 1, -1),
      createShadow(0, 1, 1, 0, 0, 1, 1, 0, 0, 1, 1, -1),
      createShadow(0, 1, 1, 0, 0, 1, 1, 0, 0, 1, 1, -1),
      createShadow(0, 1, 1, 0, 0, 1, 1, 0, 0, 1, 1, -1),
      createShadow(0, 1, 1, 0, 0, 1, 1, 0, 0, 1, 1, -1),
      createShadow(0, 1, 1, 0, 0, 1, 1, 0, 0, 1, 1, -1),
      createShadow(0, 1, 1, 0, 0, 1, 1, 0, 0, 1, 1, -1),
      createShadow(0, 1, 1, 0, 0, 1, 1, 0, 0, 1, 1, -1),
      createShadow(0, 1, 1, 0, 0, 1, 1, 0, 0, 1, 1, -1),
      createShadow(0, 1, 1, 0, 0, 1, 1, 0, 0, 1, 1, -1),
      createShadow(0, 2, 4, -1, 0, 4, 5, 0, 0, 1, 10, 0),
      createShadow(0, 3, 5, -1, 0, 5, 8, 0, 0, 1, 14, 0),
      createShadow(0, 3, 5, -1, 0, 6, 10, 0, 0, 1, 18, 0),
      createShadow(0, 4, 5, -2, 0, 7, 10, 1, 0, 2, 16, 1),
      createShadow(0, 5, 5, -3, 0, 8, 10, 1, 0, 3, 14, 2),
      createShadow(0, 5, 6, -3, 0, 9, 12, 1, 0, 3, 16, 2),
      createShadow(0, 6, 6, -3, 0, 10, 14, 1, 0, 4, 18, 3),
      createShadow(0, 6, 7, -4, 0, 11, 15, 1, 0, 4, 20, 3),
    ],
  });

export default dva;
