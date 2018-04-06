import React from "react";
import SvgIcon from "material-ui/SvgIcon";

class Icon extends React.PureComponent {
  render() {
    return (
      <SvgIcon {...this.props}>
        <path
          fillRule="evenodd"
          d="M12 19.5c6.08 0 11-.87 11-7.5s-4.92-7.5-11-7.5S1 5.37 1 12s4.92 7.5 11 7.5zM9.08 8.3a1 1 0 0 1 .45-.25l.95-.3.6 1.9-.75.24v.27c0 .48.02.5 1.67.5 1.64 0 1.66-.02 1.66-.5v-.27l-.71-.22.57-1.91.96.29c.19.05.29.12.42.23.57.46.76 1.05.76 1.88 0 1.83-1.05 2.5-3.66 2.5-2.61 0-3.67-.67-3.67-2.5 0-.8.18-1.4.74-1.84l.01-.02zm1.3 1.53z"
        />
      </SvgIcon>
    );
  }
}

export default Icon;
