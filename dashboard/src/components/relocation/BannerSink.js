import React from "react";
import PropTypes from "prop-types";

import { Sink } from "/components/relocation/Relocation";
import { BANNER } from "/components/relocation/types";

class BannerSink extends React.PureComponent {
  static propTypes = {
    children: PropTypes.node.isRequired,
  };

  render() {
    return <Sink>{{ render: () => this.props.children, type: BANNER }}</Sink>;
  }
}

export default BannerSink;
