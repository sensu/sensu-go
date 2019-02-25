import React from "react";
import PropTypes from "prop-types";

import uniqueId from "/lib/util/uniqueId";

class AnimatedLogo extends React.PureComponent {
  static propTypes = {
    style: PropTypes.object,
    size: PropTypes.number,
    color: PropTypes.string,
  };

  static defaultProps = {
    style: {},
    size: 100,
    color: "currentColor",
  };

  componentWillMount() {
    this._id = `AnimatedLogo-${uniqueId()}`;
  }

  render() {
    const { style, size, color } = this.props;

    const edge = 100 / 1.414;
    const strokeWidth = 0.14 * edge;

    const circle = animation => (
      <circle cx={0} cy={0} clipPath={`url(#${this._id}-innerClipPath)`}>
        <animate
          attributeType="XML"
          attributeName="r"
          from={strokeWidth / 2}
          to={100}
          dur="2s"
          repeatCount="indefinite"
          calcMode="spline"
          keySplines="0.5 0 1 0.5"
          {...animation}
        />
      </circle>
    );

    return (
      <svg
        style={style}
        width={size}
        height={size}
        viewBox={`0 0 100 100`}
        fill="none"
      >
        <defs>
          <rect
            id={`${this._id}-outerRect`}
            x={0}
            y={0}
            width={edge}
            height={edge}
          />
          <clipPath id={`${this._id}-outerClipPath`}>
            <use href={`#${this._id}-outerRect`} />
          </clipPath>
          <rect
            id={`${this._id}-innerRect`}
            x={strokeWidth}
            y={strokeWidth}
            width={edge - strokeWidth * 2}
            height={edge - strokeWidth * 2}
          />
          <clipPath id={`${this._id}-innerClipPath`}>
            <use href={`#${this._id}-innerRect`} />
          </clipPath>
        </defs>
        <g
          transform={[
            `rotate(225 50 50)`,
            `translate(${50 - edge / 2} ${50 - edge / 2})`,
          ]}
          stroke={color}
          strokeWidth={strokeWidth}
        >
          <use
            href={`#${this._id}-outerRect`}
            strokeWidth={strokeWidth * 2 + 0.5}
            clipPath={`url(#${this._id}-outerClipPath)`}
            fill="none"
          />
          {circle()}
          {circle({ begin: `-${2 / 3}s` })}
          {circle({ begin: `-${4 / 3}s` })}
        </g>
      </svg>
    );
  }
}

export default AnimatedLogo;
