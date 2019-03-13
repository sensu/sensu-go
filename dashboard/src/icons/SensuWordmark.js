import React from "react";
import PropTypes from "prop-types";
import { compose, pure } from "recompose";
import SvgIcon from "@material-ui/core/SvgIcon";
import classnames from "classnames";
import { withStyles } from "@material-ui/core/styles";

class Icon extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    className: PropTypes.string,
    viewBox: PropTypes.string,
  };

  static defaultProps = {
    className: "",
    viewBox: "0 0 523 123",
  };

  static styles = {
    root: {
      width: "initial",
    },
  };

  render() {
    const { className: classNameProp, viewBox, classes, ...props } = this.props;
    const className = classnames(classes.root, classNameProp);

    return (
      <SvgIcon {...props} viewBox={viewBox} className={className}>
        <path
          fillRule="evenodd"
          d="M142 68c2-12 8-18 18-18 9 0 15 6 16 18h-34zm18-39c-15 0-27 4-35 12s-12 20-12 35 4 26 12 34c9 8 21 12 37 12 10 0 19-2 26-6s12-9 15-17l-25-8c-1 3-3 6-6 7-2 2-6 3-10 3-6 0-10-1-13-4s-5-7-6-12h61l1-11c0-14-4-25-12-33-7-8-19-12-33-12zm116 0c-7 0-13 2-19 5-5 3-9 8-12 14l-1-17h-27v90h31V73c0-7 1-12 4-16 3-3 7-5 12-5 4 0 7 2 9 4 2 3 3 7 3 13v52h31V61c0-10-3-18-8-24-6-5-14-8-23-8m77 22l11-2c5 0 10 1 14 3s7 5 10 9l16-15c-4-6-10-10-16-13-7-3-15-4-25-4s-18 1-24 4c-7 3-12 7-15 11-3 5-5 10-5 15 0 7 3 13 9 18s15 8 29 11l12 3c3 2 4 3 4 5s-1 3-4 4c-2 2-5 2-10 2-12 0-21-3-27-11l-15 16c9 10 23 15 44 15 14 0 24-2 32-8 7-5 11-12 11-21 0-8-3-14-9-18-6-5-15-8-29-11l-12-3c-3-2-4-3-4-5s1-3 3-5m160 50c-3 0-5-1-7-3l-1-8V31h-31v53c0 4-2 8-4 10-3 4-7 5-12 5-4 0-7-1-9-3-2-3-3-7-3-13V31h-31v60c0 11 3 19 8 24 6 5 13 8 23 8 14 0 24-6 30-16 1 3 2 7 5 9 4 4 11 7 20 7a44 44 0 0 0 19-5l2-19-9 2M41 27c3-2 8-3 14-3 7 0 14 1 19 4l16 12 16-19C99 14 91 8 83 5a79 79 0 0 0-56 0c-8 3-14 8-18 14S3 31 3 39c0 7 2 13 6 18s9 9 15 11c7 3 15 6 26 8 8 1 14 3 17 5 4 1 6 4 6 6 0 4-2 6-5 8-4 2-9 3-15 3-8 0-15-1-21-4-6-2-11-6-17-12L0 102c6 7 13 12 22 15a91 91 0 0 0 58 1c8-3 14-7 18-13 4-5 7-12 7-19 0-11-4-19-11-25-7-5-19-10-35-13-9-1-15-3-18-5-4-2-5-4-5-7 0-4 1-7 5-9"
        />
      </SvgIcon>
    );
  }
}

const EnhancedIcon = compose(
  pure,
  withStyles(Icon.styles),
)(Icon);
EnhancedIcon.muiName = "SvgIcon";

export default EnhancedIcon;
