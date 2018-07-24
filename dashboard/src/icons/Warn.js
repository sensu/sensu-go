import React from "react";
import PropTypes from "prop-types";
import SvgIcon from "@material-ui/core/SvgIcon";

let id = 0;
const getNextId = () => {
  id += 1;
  return id;
};

class Icon extends React.PureComponent {
  static propTypes = {
    withGap: PropTypes.bool,
  };

  static defaultProps = {
    withGap: false,
  };

  componentWillMount() {
    this._id = `AnimatedLogo-${getNextId()}`;
  }

  render() {
    const { withGap, ...props } = this.props;
    const gapIdx = `ic-err-${this._id}-gap`;

    return (
      <SvgIcon {...props}>
        {withGap && (
          <defs>
            <mask id={gapIdx}>
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
        )}

        <g fill="currentColor" mask={withGap && `url(#${gapIdx})`}>
          <path d="M1 21h22L12 2 1 21zm12-3h-2v-2h2v2zm0-4h-2v-4h2v4z" />
        </g>
      </SvgIcon>
    );
  }
}

export default Icon;
