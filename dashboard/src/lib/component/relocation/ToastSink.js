import React from "react";
import PropTypes from "prop-types";

import { Sink } from "/lib/component/relocation/Relocation";
import { TOAST } from "/lib/component/relocation/types";

class ToastSink extends React.PureComponent {
  static propTypes = {
    children: PropTypes.node.isRequired,
  };

  render() {
    return <Sink>{{ render: () => this.props.children, type: TOAST }}</Sink>;
  }
}

export default ToastSink;
