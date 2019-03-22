import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "@material-ui/core/styles";
import SvgIcon from "@material-ui/core/SvgIcon";

// https://material.io/design/iconography/animated-icons.html#transitions
const duration = 500;
const outDuration = 100;

const styles = {
  root: {},
  inactive: {
    "& $radio": {
      opacity: 1,
      fill: "none",
      transform: "translateY(0) scale(1)",
      transition: `
        opacity ${outDuration}ms ease-in,
        transform ${outDuration}ms ease-in
      `,
    },
    "& $antenna": {
      opacity: 0,
      transition: "none",
    },
    "& $waves": {
      "& path": {
        opacity: 0,
        animation: "none",
        transition: "none",
      },
    },
  },
  waves: {
    "& path": {
      opacity: 1,
      transition: `opacity ${duration}ms ease-out`,
      transitionDelay: duration * (9 / 16),
      fillOpacity: 0.5,
      animation: "10s ease-in-out 2s normal infinite ic-live-gentle-ripple",
    },
    // stagger transition & animations to give an impression of a ripple
    "& path:last-child": {
      transitionDelay: duration * (13 / 16),
      animationDelay: "3s",
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
  "@keyframes ic-live-gentle-ripple": {
    "0%": {
      fillOpacity: 0.5,
    },
    "10%": {
      fillOpacity: 0.95,
    },
    "20%,100%": {
      fillOpacity: 0.5,
    },
  },
};

class Icon extends React.PureComponent {
  static propTypes = {
    active: PropTypes.bool,
    classes: PropTypes.object.isRequired,
    className: PropTypes.string,
  };

  static defaultProps = {
    active: true,
    className: undefined,
  };

  render() {
    const { classes, className: classNameProp, active, ...props } = this.props;
    const className = classnames(classNameProp, {
      [classes.inactive]: !active,
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
          <path d="M16.95 4.05l-1.41 1.42C16.5 6.46 17 7.64 17 9c0 1.36-.49 2.54-1.46 3.54l1.41 1.41a6.97 6.97 0 0 0 0-9.9zM7.05 4.05l1.41 1.41a4.98 4.98 0 0 0 0 7.08l-1.41 1.41A6.92 6.92 0 0 1 5 9c0-1.88.68-3.53 2.05-4.95z" />
          <path d="M4.22 1.23l1.42 1.4A9.08 9.08 0 0 0 3 9c0 2.37.88 4.49 2.64 6.36l-1.42 1.42A10.8 10.8 0 0 1 1 9c0-2.97 1.08-5.56 3.22-7.77zM19.78 1.23l-1.41 1.4A9.08 9.08 0 0 1 21.01 9c0 2.37-.88 4.49-2.64 6.36l1.41 1.42A10.8 10.8 0 0 0 23 9c0-2.97-1.07-5.56-3.22-7.77z" />
        </g>
      </SvgIcon>
    );
  }
}

export default withStyles(styles)(Icon);
