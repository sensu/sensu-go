import React from "react";
import pure from "recompose/pure";
import SvgIcon from "material-ui/SvgIcon";

class Icon extends React.Component {
  render() {
    return (
      <SvgIcon {...this.props}>
        <path
          d="M17 16v3c0 .7-.3 1-1 1H5c-.7 0-1-.3-1-1V5c0-.7.3-1 1-1h11c.7 0 1 .3 1 1v3c.3-.5 1-.8 2-.8s1.8.5 2.3 1.5a23 23 0 0 1 0 6.6c-.6 1-1.3 1.4-2.3 1.4-1 0-1.7-.2-2-.7zm2-7a1 1 0 0 0-1 1v4c0 .6.4 1 1 1s1-.4 1-1a22.6 22.6 0 0 0 0-4c0-.6-.4-1-1-1zM7 10.2c0 1.3.9 2.7 2.8 4.3.7.6.7.6 1.4 0 2-1.6 2.8-3 2.8-4.3 0-.9-.9-1.4-1.8-1.4-.4 0-.8.3-1.3.9-.4.3-.4.3-.8 0-.5-.6-.9-.9-1.3-.9-1 0-1.8.5-1.8 1.4z"
          fillRule="evenodd"
        />
      </SvgIcon>
    );
  }
}

const HeartMugIcon = pure(Icon);
HeartMugIcon.muiName = "SvgIcon";

export default HeartMugIcon;
