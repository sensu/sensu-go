import React from "react";
import PropTypes from "prop-types";

import uniqueId from "/utils/uniqueId";

class IconGap extends React.PureComponent {
  static propTypes = {
    children: PropTypes.func.isRequired,
    disabled: PropTypes.bool.isRequired,
  };

  componentWillMount() {
    if (!this.props.disabled) {
      this._id = `ic-gap-${uniqueId()}`;
    }
  }

  render() {
    const { children, disabled } = this.props;
    const maskId = `${this._id}-mask`;

    if (disabled) {
      return children({ maskId: "" });
    }

    return (
      <React.Fragment>
        <defs>
          <mask id={maskId}>
            <rect x="0" y="0" width="24" height="24" fill="white" />
            <rect
              x="12"
              y="12"
              width="12"
              height="12"
              rx="2"
              ry="2"
              fill="black"
            />
          </mask>
        </defs>

        {children({ maskId })}
      </React.Fragment>
    );
  }
}

export default IconGap;
