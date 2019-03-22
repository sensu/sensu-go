import React from "react";
import PropTypes from "prop-types";

import { Consumer } from "/lib/component/relocation/Relocation";
import { BANNER } from "/lib/component/relocation/types";

class BannerConnector extends React.PureComponent {
  static propTypes = {
    children: PropTypes.func.isRequired,
  };

  render() {
    return (
      <Consumer>
        {({ createChild }) =>
          this.props.children({
            addBanner: render => createChild({ render, type: BANNER }),
          })
        }
      </Consumer>
    );
  }
}

export default BannerConnector;
