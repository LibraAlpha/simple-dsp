<template>
  <div class="bid-management">
    <div class="page-header">
      <h2>出价管理</h2>
      <el-button type="primary" @click="handleAdd">
        <el-icon><Plus /></el-icon>
        添加出价策略
      </el-button>
    </div>

    <el-card class="filter-card">
      <el-form :inline="true" :model="filterForm" class="filter-form">
        <el-form-item label="计费类型">
          <el-select v-model="filterForm.bidType" placeholder="请选择">
            <el-option label="全部" value="" />
            <el-option label="CPC" value="CPC" />
            <el-option label="CPM" value="CPM" />
          </el-select>
        </el-form-item>
        <el-form-item label="出价范围">
          <el-input-number
            v-model="filterForm.minPrice"
            :precision="4"
            :step="0.0001"
            placeholder="最小值"
          />
          <span class="separator">-</span>
          <el-input-number
            v-model="filterForm.maxPrice"
            :precision="4"
            :step="0.0001"
            placeholder="最大值"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleFilter">查询</el-button>
          <el-button @click="resetFilter">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="bid-table">
      <el-table
        v-loading="loading"
        :data="bidList"
        border
        style="width: 100%"
      >
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="策略名称" min-width="150" />
        <el-table-column prop="bidType" label="计费类型" width="100">
          <template #default="{ row }">
            <el-tag :type="row.bidType === 'CPC' ? 'success' : 'warning'">
              {{ row.bidType }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="price" label="出价" width="150">
          <template #default="{ row }">
            {{ formatPrice(row.price, row.bidType) }}
            <span class="unit">{{ row.bidType === 'CPC' ? '元' : '分' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="dailyBudget" label="日预算" width="150">
          <template #default="{ row }">
            {{ formatMoney(row.dailyBudget) }} 元
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-switch
              v-model="row.status"
              :active-value="1"
              :inactive-value="0"
              @change="handleStatusChange(row)"
            />
          </template>
        </el-table-column>
        <el-table-column label="关联素材" min-width="200">
          <template #default="{ row }">
            <el-tag
              v-for="creative in row.creatives"
              :key="creative.id"
              class="creative-tag"
              closable
              @close="handleRemoveCreative(row, creative)"
            >
              素材ID: {{ creative.creativeId }}
            </el-tag>
            <el-button
              type="primary"
              link
              @click="handleAddCreative(row)"
            >
              添加素材
            </el-button>
          </template>
        </el-table-column>
        <el-table-column prop="updatedAt" label="更新时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.updatedAt) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button-group>
              <el-button type="primary" link @click="handleEdit(row)">
                编辑
              </el-button>
              <el-button type="primary" link @click="handleViewStats(row)">
                统计
              </el-button>
              <el-button type="danger" link @click="handleDelete(row)">
                删除
              </el-button>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <!-- 出价策略编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogType === 'add' ? '添加出价策略' : '编辑出价策略'"
      width="600px"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="策略名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入策略名称" />
        </el-form-item>
        <el-form-item label="计费类型" prop="bidType">
          <el-select 
            v-model="form.bidType" 
            placeholder="请选择计费类型"
            :disabled="dialogType === 'edit'"
          >
            <el-option label="CPC" value="CPC" />
            <el-option label="CPM" value="CPM" />
          </el-select>
        </el-form-item>
        <el-form-item label="出价" prop="price">
          <el-input-number
            v-model="form.price"
            :precision="form.bidType === 'CPC' ? 4 : 0"
            :step="form.bidType === 'CPC' ? 0.0001 : 1"
            :min="0"
            :disabled="dialogType === 'edit' && form.isPriceLocked === 1"
            style="width: 200px"
          />
          <span class="unit">{{ form.bidType === 'CPC' ? '元' : '分' }}</span>
        </el-form-item>
        <el-form-item label="日预算" prop="dailyBudget">
          <el-input-number
            v-model="form.dailyBudget"
            :precision="2"
            :step="100"
            :min="0"
            style="width: 200px"
          />
          <span class="unit">元</span>
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-switch
            v-model="form.status"
            :active-value="1"
            :inactive-value="0"
          />
        </el-form-item>
        <el-form-item v-if="dialogType === 'add'" label="锁定出价" prop="isPriceLocked">
          <el-switch
            v-model="form.isPriceLocked"
            :active-value="1"
            :inactive-value="0"
          />
          <span class="tip">锁定后将无法修改出价</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 添加素材对话框 -->
    <el-dialog
      v-model="creativeDialogVisible"
      title="添加素材"
      width="500px"
    >
      <el-form
        ref="creativeFormRef"
        :model="creativeForm"
        :rules="creativeRules"
        label-width="100px"
      >
        <el-form-item label="素材ID" prop="creativeId">
          <el-input-number
            v-model="creativeForm.creativeId"
            :min="1"
            style="width: 200px"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="creativeDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCreativeSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 统计数据对话框 -->
    <el-dialog
      v-model="statsDialogVisible"
      title="统计数据"
      width="800px"
    >
      <el-form :inline="true" :model="statsForm" class="stats-form">
        <el-form-item label="日期范围">
          <el-date-picker
            v-model="statsForm.dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadStats">查询</el-button>
        </el-form-item>
      </el-form>

      <el-table
        v-loading="statsLoading"
        :data="statsList"
        border
        style="width: 100%"
      >
        <el-table-column prop="date" label="日期" width="120" />
        <el-table-column prop="creativeId" label="素材ID" width="100" />
        <el-table-column prop="impressions" label="展示量" />
        <el-table-column prop="clicks" label="点击量" />
        <el-table-column prop="spend" label="花费">
          <template #default="{ row }">
            {{ formatMoney(row.spend) }} 元
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import dayjs from 'dayjs'
import { useBidStore } from '@/stores/bid'

const bidStore = useBidStore()
const loading = ref(false)
const bidList = ref([])
const dialogVisible = ref(false)
const dialogType = ref('add')
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

// 筛选表单
const filterForm = ref({
  bidType: '',
  minPrice: null,
  maxPrice: null
})

// 编辑表单
const form = ref({
  name: '',
  bidType: 'CPC',
  price: 0,
  dailyBudget: 0,
  status: 1,
  isPriceLocked: 1
})

const rules = {
  name: [
    { required: true, message: '请输入策略名称', trigger: 'blur' },
    { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
  ],
  bidType: [
    { required: true, message: '请选择计费类型', trigger: 'change' }
  ],
  price: [
    { required: true, message: '请输入出价', trigger: 'blur' },
    { type: 'number', min: 0, message: '出价必须大于0', trigger: 'blur' }
  ],
  dailyBudget: [
    { required: true, message: '请输入日预算', trigger: 'blur' },
    { type: 'number', min: 0, message: '日预算必须大于0', trigger: 'blur' }
  ]
}

// 素材相关
const creativeDialogVisible = ref(false)
const creativeForm = ref({
  strategyId: null,
  creativeId: null
})

const creativeRules = {
  creativeId: [
    { required: true, message: '请输入素材ID', trigger: 'blur' },
    { type: 'number', min: 1, message: '素材ID必须大于0', trigger: 'blur' }
  ]
}

// 统计相关
const statsDialogVisible = ref(false)
const statsLoading = ref(false)
const statsList = ref([])
const statsForm = ref({
  strategyId: null,
  dateRange: []
})

const formRef = ref(null)
const creativeFormRef = ref(null)

// 格式化价格
const formatPrice = (price, type) => {
  if (type === 'CPC') {
    return price.toFixed(4)
  }
  return Math.round(price)
}

// 格式化金额
const formatMoney = (amount) => {
  return amount.toFixed(2)
}

// 格式化时间
const formatTime = (time) => {
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

// 加载出价列表
const loadBidList = async () => {
  loading.value = true
  try {
    const { list, total: totalCount } = await bidStore.getBidList({
      page: page.value,
      pageSize: pageSize.value,
      ...filterForm.value
    })
    bidList.value = list
    total.value = totalCount
  } catch (error) {
    ElMessage.error('加载出价列表失败')
  } finally {
    loading.value = false
  }
}

// 处理筛选
const handleFilter = () => {
  page.value = 1
  loadBidList()
}

// 重置筛选
const resetFilter = () => {
  filterForm.value = {
    bidType: '',
    minPrice: null,
    maxPrice: null
  }
  handleFilter()
}

// 处理分页
const handleSizeChange = (val) => {
  pageSize.value = val
  loadBidList()
}

const handleCurrentChange = (val) => {
  page.value = val
  loadBidList()
}

// 添加出价策略
const handleAdd = () => {
  dialogType.value = 'add'
  form.value = {
    name: '',
    bidType: 'CPC',
    price: 0,
    dailyBudget: 0,
    status: 1,
    isPriceLocked: 1
  }
  dialogVisible.value = true
}

// 编辑出价策略
const handleEdit = (row) => {
  dialogType.value = 'edit'
  form.value = { ...row }
  dialogVisible.value = true
}

// 删除出价策略
const handleDelete = (row) => {
  ElMessageBox.confirm(
    '确定要删除该出价策略吗？',
    '提示',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(async () => {
    try {
      await bidStore.deleteBid(row.id)
      ElMessage.success('删除成功')
      loadBidList()
    } catch (error) {
      ElMessage.error('删除失败')
    }
  })
}

// 修改状态
const handleStatusChange = async (row) => {
  try {
    await bidStore.updateBidStatus(row.id, row.status)
    ElMessage.success('状态更新成功')
  } catch (error) {
    ElMessage.error('状态更新失败')
    row.status = row.status === 1 ? 0 : 1 // 恢复状态
  }
}

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      try {
        if (dialogType.value === 'add') {
          await bidStore.addBid(form.value)
          ElMessage.success('添加成功')
        } else {
          await bidStore.updateBid(form.value)
          ElMessage.success('更新成功')
        }
        dialogVisible.value = false
        loadBidList()
      } catch (error) {
        ElMessage.error(dialogType.value === 'add' ? '添加失败' : '更新失败')
      }
    }
  })
}

// 添加素材
const handleAddCreative = (row) => {
  creativeForm.value = {
    strategyId: row.id,
    creativeId: null
  }
  creativeDialogVisible.value = true
}

// 移除素材
const handleRemoveCreative = async (strategy, creative) => {
  try {
    await bidStore.removeCreative(strategy.id, creative.creativeId)
    ElMessage.success('移除素材成功')
    loadBidList()
  } catch (error) {
    ElMessage.error('移除素材失败')
  }
}

// 提交素材表单
const handleCreativeSubmit = async () => {
  if (!creativeFormRef.value) return

  await creativeFormRef.value.validate(async (valid) => {
    if (valid) {
      try {
        await bidStore.addCreative(
          creativeForm.value.strategyId,
          creativeForm.value.creativeId
        )
        ElMessage.success('添加素材成功')
        creativeDialogVisible.value = false
        loadBidList()
      } catch (error) {
        ElMessage.error('添加素材失败')
      }
    }
  })
}

// 查看统计数据
const handleViewStats = (row) => {
  statsForm.value = {
    strategyId: row.id,
    dateRange: [
      dayjs().subtract(7, 'day').format('YYYY-MM-DD'),
      dayjs().format('YYYY-MM-DD')
    ]
  }
  statsDialogVisible.value = true
  loadStats()
}

// 加载统计数据
const loadStats = async () => {
  if (!statsForm.value.dateRange || statsForm.value.dateRange.length !== 2) {
    ElMessage.warning('请选择日期范围')
    return
  }

  statsLoading.value = true
  try {
    const stats = await bidStore.getStrategyStats(
      statsForm.value.strategyId,
      statsForm.value.dateRange[0],
      statsForm.value.dateRange[1]
    )
    statsList.value = stats
  } catch (error) {
    ElMessage.error('加载统计数据失败')
  } finally {
    statsLoading.value = false
  }
}

onMounted(() => {
  loadBidList()
})
</script>

<style lang="scss" scoped>
.bid-management {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    
    h2 {
      margin: 0;
      font-weight: 500;
    }
  }
  
  .filter-card {
    margin-bottom: 20px;
    
    .filter-form {
      .separator {
        margin: 0 10px;
      }
    }
  }
  
  .bid-table {
    .unit {
      margin-left: 4px;
      color: #909399;
      font-size: 12px;
    }
    
    .pagination {
      margin-top: 20px;
      text-align: right;
    }

    .creative-tag {
      margin-right: 8px;
      margin-bottom: 8px;
    }
  }

  .tip {
    margin-left: 8px;
    color: #909399;
    font-size: 12px;
  }

  .stats-form {
    margin-bottom: 20px;
  }
}
</style> 