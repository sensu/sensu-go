import React from "react";
import PropTypes from "prop-types";

import Button from "@material-ui/core/Button";
import ButtonIcon from "/components/ButtonIcon";
import LiveIcon from "/icons/Live";

class LiveButton extends React.PureComponent {
  static propTypes = {
    active: PropTypes.bool.isRequired,
  };

  render() {
    const { active, ...props } = this.props;
    return (
      <Button {...props}>
        <ButtonIcon>
          <LiveIcon active={active} />
        </ButtonIcon>
        LIVE
      </Button>
    );
  }
}

export default LiveButton;
