import React from "react";
import pure from "recompose/pure";
import SvgIcon from "@material-ui/core/SvgIcon";

class Icon extends React.Component {
  render() {
    return (
      <SvgIcon {...this.props}>
        <g fillRule="evenodd">
          <path d="M17 12.8l2.2-.6c1.3-.4 2.5.4 2.5 1.8 0 1.9-2.7 4.2-8 7h-.1c-1 .4-2.2.5-3.6.5-4.6 0-7-1.5-7-9.9h14v1.2zm-.3 4.6c2-1 2.9-2 2.9-2.7 0-.8-.5-1.2-1.4-.8l-1.1.3c0 .5 0 1-.2 1.6l-.2 1.6z" />
          <path
            d="M9.9 2c.3-.2.5-.2.7 0 .2.2.2.4 0 .7-.4.5-.3 1.2-.3 1.6l.8 1.5c.7 1.2.7 1.6.7 2 0 .7-.5 1.5-1.2 2.1-.2.3-.5.3-.7.1-.2-.2-.2-.4 0-.7.5-.5.3-1 .3-1.4 0-.3-.3-.6-.8-1.5-.6-1.1-.7-1.6-.7-2 0-.8.6-1.6 1.2-2.4z"
            fillRule="nonzero"
          />
        </g>
      </SvgIcon>
    );
  }
}

const EspressoIcon = pure(Icon);
EspressoIcon.muiName = "SvgIcon";

export default EspressoIcon;
