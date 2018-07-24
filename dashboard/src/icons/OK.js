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
          <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z" />
        </g>
      </SvgIcon>
    );
  }
}

export default Icon;
