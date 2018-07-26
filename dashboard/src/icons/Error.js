import React from "react";
import PropTypes from "prop-types";
import SvgIcon from "@material-ui/core/SvgIcon";
import IconGap from "./IconGap";

class Icon extends React.PureComponent {
  static propTypes = {
    withGap: PropTypes.bool,
  };

  static defaultProps = {
    withGap: false,
  };

  render() {
    const { withGap, ...props } = this.props;

    return (
      <SvgIcon {...props}>
        <IconGap disabled={!withGap}>
          {({ maskId }) => (
            <g fill="none" fillRule="evenodd" mask={`url(#${maskId})`}>
              <path d="M0 0h24v24H0z" />
              <path
                d="M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"
                fill="currentColor"
                fillRule="nonzero"
              />
            </g>
          )}
        </IconGap>
      </SvgIcon>
    );
  }
}

export default Icon;
