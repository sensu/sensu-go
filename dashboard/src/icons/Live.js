import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "@material-ui/core/styles";
import SvgIcon from "@material-ui/core/SvgIcon";

const duration = 300;
const styles = {
  root: {},
  inactive: {
    "& $radio": {
      opacity: 1,
      fill: "none",
      transform: "translateY(0) scale(1)",
      transition: "opacity 100ms ease-in, transform 100ms ease-in",
    },
    "& $antenna": {
      opacity: 0,
      transition: "none",
    },
    "& $waves": {
      opacity: 0,
      transform: "scale(0.8)",
      transition: "none",
      "& path": {
        animation: "none",
      },
    },
  },
  waves: {
    transition: `transform ${duration}ms ease, opacity ${duration}ms`,
    transitionDelay: duration * (5 / 8),
    transformOrigin: "center",
    "& path": {
      fillOpacity: 0.66,
      animation: "10s ease-in-out 0s normal infinite live-gentle-pulse",
    },
    "& path:last-child": {
      animationDelay: "1s",
    },
  },
  antenna: {
    transition: `opacity ${duration}ms`,
    transitionDelay: duration * (3 / 8),
  },
  radio: {
    opacity: 0,
    fill: "currentColor",
    transform: "translateY(-13%) scale(0.4)",
    transformOrigin: "center",
    transition: `
      transform ${duration * (2 / 3)}ms ease,
      opacity ${duration * (2 / 3)}ms step-end
    `,
  },
  "@keyframes live-gentle-pulse": {
    "0%": {
      fillOpacity: 0.66,
    },
    "10%": {
      fillOpacity: 0.95,
    },
    "20%,100%": {
      fillOpacity: 0.66,
    },
  },
};

class Icon extends React.PureComponent {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    className: PropTypes.string,
    inactive: PropTypes.bool,
  };

  static defaultProps = {
    inactive: false,
    className: undefined,
  };

  render() {
    const {
      classes,
      className: classNameProp,
      inactive,
      ...props
    } = this.props;
    const className = classnames(classNameProp, {
      [classes.inactive]: inactive,
    });

    return (
      <SvgIcon className={className} {...props}>
        <circle
          stroke="currentColor"
          strokeWidth="2"
          cx="12"
          cy="12"
          r="6"
          fill="none"
          fillRule="evenodd"
          className={classes.radio}
        />
        <g fillRule="evenodd" className={classes.antenna}>
          <path d="M11.08 11.86a3 3 0 1 1 1.84 0l.95 7.14c0 .67-.29 1-.87 1h-2c-.58 0-.87-.33-.87-1l.95-7.14z" />
        </g>
        <g fillRule="evenodd" className={classes.waves}>
          <path d="M16.26 3.44l-1.21 1.6A5.3 5.3 0 0 1 17 9a5.2 5.2 0 0 1-1.95 4l1.21 1.57A6.94 6.94 0 0 0 19 9c0-2.21-.91-4.07-2.74-5.56zM7.74 3.44l1.21 1.6A5.3 5.3 0 0 0 7 9a5.2 5.2 0 0 0 1.95 4l-1.2 1.57A6.94 6.94 0 0 1 5 9c0-2.21.91-4.07 2.74-5.56z" />
          <path d="M18.8.2l-1.2 1.57A8.81 8.81 0 0 1 21 9c0 3-1.13 5.42-3.4 7.27l1.2 1.56C21.6 15.61 23 12.67 23 9S21.6 2.4 18.8.2zM5.2.2l1.2 1.57A8.81 8.81 0 0 0 3 9c0 3 1.14 5.42 3.4 7.27l-1.2 1.56A10.71 10.71 0 0 1 1 9C1 5.33 2.4 2.4 5.2.2z" />
        </g>
      </SvgIcon>
    );
  }
}

export default withStyles(styles)(Icon);
