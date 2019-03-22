import React from "react";
import pure from "recompose/pure";
import SvgIcon from "@material-ui/core/SvgIcon";

class Icon extends React.Component {
  render() {
    return (
      <SvgIcon {...this.props}>
        <g fillRule="evenodd">
          <path d="M22.5 12.5l-4.8 8.4a1 1 0 0 1-1 .6H7.1a1 1 0 0 1-1-.6l-4.8-8.4a1 1 0 0 1 0-1.1L6.1 3a1 1 0 0 1 1-.6h9.6c.4 0 .8.2 1 .6l4.8 8.4c.2.3.2.8 0 1.1zm-15-.5c0 2.4 2 4.4 4.4 4.4a4.5 4.5 0 0 0 0-8.9c-2.4 0-4.4 2-4.4 4.5z" />
          <ellipse opacity=".25" cx="11.9" cy="12" rx="4.4" ry="4.4" />
        </g>
      </SvgIcon>
    );
  }
}

const PolyIcon = pure(Icon);
PolyIcon.muiName = "SvgIcon";

export default PolyIcon;
