/**
 * errors.ts - proto-tester 自定义错误类层次
 *
 * 职责：
 * - 定义与业务码对应的异常类型
 * - 携带 code / businessCode / traceId 等诊断信息
 *
 * 错误码映射（来自 go/common-lib/codes.go）：
 * - 10200 → 成功（不抛异常）
 * - 10400 → BadRequestError
 * - 10401 → UnauthorizedError（Token 过期）
 * - 10404 → NotFoundError
 * - 10500 → InternalError
 * - 10501 → NotImplementedError
 */

/** 业务码常量 */
export const CODE_SUCCESS = 10200;
export const CODE_BAD_REQUEST = 10400;
export const CODE_UNAUTHORIZED = 10401;
export const CODE_NOT_FOUND = 10404;
export const CODE_INTERNAL_ERROR = 10500;
export const CODE_NOT_IMPLEMENTED = 10501;

/** proto-tester 基础错误类 */
export class ProtoTesterError extends Error {
  constructor(
    public code: number,
    public businessCode?: number,
    public traceId?: string,
    message?: string,
  ) {
    super(message || `ProtoTesterError(code=${code}, businessCode=${businessCode})`);
    this.name = 'ProtoTesterError';
  }
}

/** 400 参数错误 */
export class BadRequestError extends ProtoTesterError {
  constructor(businessCode?: number, traceId?: string) {
    super(400, businessCode, traceId, `BadRequestError: 业务码 ${businessCode}`);
    this.name = 'BadRequestError';
  }
}

/** 401 未授权 / Token 过期 */
export class UnauthorizedError extends ProtoTesterError {
  constructor(businessCode?: number, traceId?: string) {
    super(401, businessCode, traceId, `UnauthorizedError: Token 已过期或无效`);
    this.name = 'UnauthorizedError';
  }
}

/** 404 未找到路由/资源 */
export class NotFoundError extends ProtoTesterError {
  constructor(businessCode?: number, traceId?: string) {
    super(404, businessCode, traceId, `NotFoundError: 业务码 ${businessCode}`);
    this.name = 'NotFoundError';
  }
}

/** 500 内部错误 */
export class InternalError extends ProtoTesterError {
  constructor(businessCode?: number, traceId?: string) {
    super(
      500,
      businessCode,
      traceId,
      `InternalError: ${traceId ? `[${traceId}] ` : ''}服务端内部错误`,
    );
    this.name = 'InternalError';
  }
}

/** 501 未实现 */
export class NotImplementedError extends ProtoTesterError {
  constructor(businessCode?: number, traceId?: string) {
    super(501, businessCode, traceId, `NotImplementedError: 业务码 ${businessCode}`);
    this.name = 'NotImplementedError';
  }
}
