import React from "react";
import PropTypes from "prop-types";
import SvgIcon from "@material-ui/core/SvgIcon";

class SensuIcon extends React.PureComponent {
  static propTypes = {
    viewBox: PropTypes.string,
  };

  static defaultProps = {
    className: "",
    viewBox: "0 0 268 268",
  };

  render() {
    const { viewBox, ...props } = this.props;
    return (
      <SvgIcon {...props} viewBox={viewBox}>
        <path
          fillRule="evenodd"
          d="M195.8 169.5a115.6 115.6 0 0 0-124.3 0L37 135a162.6 162.6 0 0 1 96.6-31.5c35.3 0 68.8 11 96.7 31.5l-34.6 34.5zM89.8 188a90.1 90.1 0 0 1 87.6 0l-43.8 43.8L89.8 188zm43.8-152.3l49.2 49.2a189.7 189.7 0 0 0-98.4 0l49.2-49.2zm133.7 98L133.6 0 0 133.6l133.6 133.7 133.7-133.7z"
        />
      </SvgIcon>
    );
  }
}

export default SensuIcon;
