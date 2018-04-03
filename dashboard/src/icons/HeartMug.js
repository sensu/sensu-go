import React from "react";
import pure from "recompose/pure";
import SvgIcon from "material-ui/SvgIcon";

class Icon extends React.Component {
  render() {
    return (
      <SvgIcon {...this.props}>
        <path
          fillRule="evenodd"
          d="M17.76 16.76v2.65c0 1.06-1.4 1.59-4.23 1.59h-5.3C5.42 21 4 20.47 4 19.41V4.6C4 3.53 5.41 3 8.24 3h5.29c2.82 0 4.23.53 4.23 1.59v3.7c.36-.53 1.06-.79 2.12-.79s1.86.53 2.38 1.59c.18 1.23.27 2.38.27 3.44s-.09 2.2-.27 3.44c-.52 1.06-1.32 1.59-2.38 1.59-1.06 0-1.76-.27-2.12-.8zm1.99-7.4a1 1 0 0 0-1 1v4.35a1 1 0 0 0 1 1h.23a1 1 0 0 0 1-.9 23.92 23.92 0 0 0 0-4.55 1 1 0 0 0-1-.9h-.23zm-12.57 2c0 1.33.94 2.9 2.96 4.58.74.61.74.61 1.48 0 2.05-1.7 2.97-3.25 2.97-4.58 0-.92-.93-1.48-1.85-1.48-.5 0-.91.36-1.49.93-.37.37-.37.37-.74 0-.57-.57-.99-.93-1.48-.93-.93 0-1.85.56-1.85 1.48zM6.12 5.65c1.06.53 8.47.53 9.53 0 1.06-.53 1.06-1.6 0-1.06-1.06.53-8.47.53-9.53 0-1.06-.53-1.06.53 0 1.06z"
        />
      </SvgIcon>
    );
  }
}

const HeartMugIcon = pure(Icon);
HeartMugIcon.muiName = "SvgIcon";

export default HeartMugIcon;
