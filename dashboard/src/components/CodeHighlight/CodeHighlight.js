import "highlight.js/styles/github.css";

import React from "react";
import PropTypes from "prop-types";
import Worker from "./CodeHighlight.worker";

class CodeHighlight extends React.Component {
  static propTypes = {
    language: PropTypes.string.isRequired,
    code: PropTypes.string.isRequired,
    children: PropTypes.node.isRequired,
  };

  state = {
    result: this.props.children,
  };

  componentDidMount() {
    const code = this.props.code;
    const worker = new Worker();
    worker.onmessage = event => {
      this.setState({ result: event.data });
    };
    worker.postMessage([this.props.language, code]);
  }

  render() {
    return this.props.children(this.state.result);
  }
}

export default CodeHighlight;
