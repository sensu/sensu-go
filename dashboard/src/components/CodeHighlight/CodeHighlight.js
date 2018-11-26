import "highlight.js/styles/github-gist.css";

import React from "react";
import PropTypes from "prop-types";
import Worker from "./CodeHighlight.worker";

const worker = new Worker();
const pending = {};

worker.onmessage = event => {
  const [key, data] = event.data;
  pending[key].call(this, data);
};

let idx = 0;
function postMessage(msg, callback) {
  const key = idx;
  pending[idx] = data => {
    callback(data);
    delete pending[idx];
  };
  worker.postMessage({ key, msg });
  idx += 1;
}

class CodeHighlight extends React.Component {
  static propTypes = {
    language: PropTypes.oneOf(["json", "bash", "properties"]).isRequired,
    code: PropTypes.string.isRequired,
    component: PropTypes.oneOfType([PropTypes.func, PropTypes.string]),
  };

  static defaultProps = {
    component: "code",
  };

  static getDerivedStateFromProps(props, state) {
    if (props.code !== state.code) {
      return { code: props.code, result: props.code, processed: false };
    }
    return null;
  }

  state = {
    code: this.props.code,
    result: this.props.code,
    processed: false,
  };

  componentDidMount() {
    this.update();
  }

  componentDidUpdate() {
    if (!this.state.processed) {
      this.update();
    }
  }

  componentWillUnmount() {
    this.unmounted = true;
  }

  update = () => {
    postMessage([this.props.language, this.props.code], result => {
      if (!this.unmounted) {
        this.setState({ result, processed: true });
      }
    });
  };

  render() {
    const { component: Component } = this.props;
    return (
      <Component dangerouslySetInnerHTML={{ __html: this.state.result }} />
    );
  }
}

export default CodeHighlight;
