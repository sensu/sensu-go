import { set, parsePath } from "@10xjs/form";

export const REQUIRED = "VALIDATION_REQUIRED";
export const UNIQUE_CONSTRAINT = "VALIDATION_UNIQUE_CONSTRAINT";

export const requiredError = () => ({
  code: REQUIRED,
});

export const uniqueConstraintError = () => ({
  code: UNIQUE_CONSTRAINT,
});

export const formatValidationError = error => {
  if (!error) {
    return "";
  }

  const messageMap = {
    [REQUIRED]: () => "Field is required.",
    [UNIQUE_CONSTRAINT]: () => "Value must be unique.",
  };

  if (messageMap[error.code]) {
    return messageMap[error.code](error);
  }

  return error.code;
};

export const isValidationError = error =>
  typeof error.input === "string" && /^VALIDATION_/.test(error.code);

export const parseValidationErrors = errors =>
  errors.reduce((result, error) => {
    if (!isValidationError(error)) {
      return result;
    }

    const { input, ...rest } = error;
    return set(result, parsePath(input), rest);
  }, {});
