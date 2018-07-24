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

        <g mask={withGap && `url(#${gapIdx})`}>
          <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 17h-2v-2h2v2zm2.07-7.75l-.9.92C13.45 12.9 13 13.5 13 15h-2v-.5c0-1.1.45-2.1 1.17-2.83l1.24-1.26c.37-.36.59-.86.59-1.41 0-1.1-.9-2-2-2s-2 .9-2 2H8c0-2.21 1.79-4 4-4s4 1.79 4 4c0 .88-.36 1.68-.93 2.25z" />
        </g>
      </SvgIcon>
    );
  }
}

export default Icon;
