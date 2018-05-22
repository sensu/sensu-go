import React from "react";
import SvgIcon from "@material-ui/core/SvgIcon";

class Icon extends React.PureComponent {
  render() {
    return (
      <SvgIcon {...this.props}>
        <g fillRule="evenodd">
          <path d="M13 14h-2v-4h2v4zm0 4h-2v-2h2v2z" />
          <path d="M12 5.5L4 19.32h16L12 5.5zM12 2L1 21h22L12 2z" />
        </g>
      </SvgIcon>
    );
  }
}

export default Icon;
