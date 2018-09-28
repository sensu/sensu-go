import React from "react";
import { pure } from "recompose";
import SvgIcon from "@material-ui/core/SvgIcon";

class Icon extends React.Component {
  render() {
    return (
      <SvgIcon {...this.props}>
        <g fillRule="evenodd">
          <path fillRule="nonzero" d="M19 13h-4V7H9v6H5l7 7z" />
          <path d="M5 4v2h14V4z" />
        </g>
      </SvgIcon>
    );
  }
}

const UnpublishIcon = pure(Icon);
UnpublishIcon.muiName = "SvgIcon";

export default UnpublishIcon;
