import React from "react";
import pure from "recompose/pure";
import SvgIcon from "material-ui/SvgIcon";

class Icon extends React.Component {
  render() {
    return (
      <SvgIcon {...this.props}>
        <path
          fillRule="nonzero"
          d="M20 6h-4V4a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v2H4c-1.11 0-1.99.89-1.99 2L2 19a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2zm-8 9a2 2 0 0 1-2-2c0-1.1.9-2 2-2a2 2 0 0 1 2 2 2 2 0 0 1-2 2zm2-9h-4V4h4v2z"
        />
      </SvgIcon>
    );
  }
}

const PureIcon = pure(Icon);
PureIcon.muiName = "SvgIcon";

export default PureIcon;
