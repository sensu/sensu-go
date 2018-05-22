import React from "react";
import SvgIcon from "@material-ui/core/SvgIcon";

class Icon extends React.PureComponent {
  render() {
    return (
      <SvgIcon {...this.props}>
        <path d="M19.48 4.85L8.75 15.58l-4.23-4.23-1.77 1.77 6 6 12.5-12.5z" />
      </SvgIcon>
    );
  }
}

export default Icon;
